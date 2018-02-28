package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// Proxies the connection and pipes the data between client and target
func doProxy(conn *serverConnection) {
	conn.status = stsRdConn
	doProcessesConnection(conn)
	doConnectToTarget(conn)

	v := make([]byte, 10)

	if conn.status != stsTgtErr {
		v[0] = 0x05 // SOCKS 5
		v[1] = 0x00 // STATUS, default to success (0), set to 0x1 if we had tgt err
		v[2] = 0x00 // always 0
		v[3] = 0x01 // IP4 type

		v[4] = conn.targetIP[0] // IP4 Octet 1
		v[5] = conn.targetIP[1] // IP4 Octet 2
		v[6] = conn.targetIP[2] // IP4 Octet 3
		v[7] = conn.targetIP[3] // IP4 Octet 4

		v[8] = 0x0 // IP4 Port
		v[9] = 0x0 // IP4 Port
	}

	if conn.status == stsTgtErr {
		v[1] = 0x1
		conn.status = stsClose
		_, err := conn.conn.Write(v)
		if err != nil {
			log.Error("Failed to write target error response")
		}
		return
	}

	_, err := conn.conn.Write(v)
	if err != nil {
		log.Error("Failed to write target header response")
	}

	// process connection
	doHandleProxying(conn)
}

// PRocesses the connection request
func doProcessesConnection(conn *serverConnection) {
	for conn.status == stsRdConn {
		cmdBuf := make([]byte, 0, 516)
		tmp := make([]byte, 4)
		for conn.status == stsRdConn {
			n, err := conn.conn.Read(tmp)
			if err != nil {
				if err == io.EOF {
					conn.status = stsClose
				} else if err != nil {
					log.Error("Failed to read", err.Error())
				}
				break
			}
			cmdBuf = append(cmdBuf, tmp[:n]...)
			if len(cmdBuf) > 4 {
				doProcessConnectionHeader(conn, cmdBuf)
			}
		}
	}
}

// Processes the connection opener header
func doProcessConnectionHeader(conn *serverConnection, cmdBuf []byte) {
	if cmdBuf[0] != 5 {
		log.Error("Version mismatch")
		conn.status = stsClose
		return
	}
	if cmdBuf[1] != 1 {
		log.Error("Only TCP/IP proxying is supported")
		conn.status = stsClose
		return
	}
	if cmdBuf[3] != 1 {
		log.Error("Only IP4 addresses are supported, want 1 got", cmdBuf[3])
		conn.status = stsClose
		return
	}

	if len(cmdBuf) == 10 {
		ip := cmdBuf[4:8]
		port := cmdBuf[8:]
		conn.status = stsTgtConn
		conn.targetIP = ip
		conn.targetPort = binary.BigEndian.Uint16(port)
	}
}

// Connects to a target server
func doConnectToTarget(conn *serverConnection) {
	if conn.targetIP == nil || len(conn.targetIP) < 4 {
		conn.status = stsTgtErr
		log.Error("IP Address is invalid", conn.targetIP)
		return
	}
	ip := net.IPv4(conn.targetIP[0], conn.targetIP[1], conn.targetIP[2], conn.targetIP[3])
	addr := strings.Join([]string{ip.String(), fmt.Sprint(conn.targetPort)}, ":")
	rConn, err := net.Dial("tcp", addr)

	if err != nil {
		conn.status = stsTgtErr
		log.Error("Failed to open remote connection to", addr)
		return
	}
	conn.status = stsProxying

	conn.targetConn = rConn
}

// Starts the actual proxying by triggerring two go routines, one to pipe from client to target, and the other vice-versa
func doHandleProxying(conn *serverConnection) {
	log.Debug("Proxying ", conn.conn.RemoteAddr(), "to", conn.targetConn.RemoteAddr())
	go egress(conn)
	go ingress(conn)

	for conn.status == stsProxying {
		time.Sleep(time.Duration(5) * time.Second)
	}
}

// egress from client to target
func egress(conn *serverConnection) {
	for conn.status == stsProxying {
		pipe(conn, conn.conn, counterWriter{
			to:   conn.targetConn,
			in:   false,
			conn: conn})
	}
}

// ingress from target to client
func ingress(conn *serverConnection) {
	for conn.status == stsProxying {
		pipe(conn, conn.targetConn, counterWriter{
			to:   conn.conn,
			in:   true,
			conn: conn})
	}
}

// 'Pipes' data between two connections by copying from one to the other
func pipe(sc *serverConnection, from io.Reader, to io.Writer) {

	for sc.status == stsProxying {
		_, err := io.Copy(to, from)
		if err != nil {
			log.Error("Failed to write data while piping")
		}
		log.Info("Closing ", sc.id)
		sc.status = stsClose
	}
}

type counterWriter struct {
	in   bool
	conn *serverConnection
	to   io.Writer
}

func (w counterWriter) Write(p []byte) (n int, err error) {
	if w.in {
		w.conn.dataIn += int64(len(p))
	} else {
		w.conn.dataOut += int64(len(p))
	}

	start := time.Now()
	written, err := w.to.Write(p)
	w.conn.activeTime += time.Now().Sub(start).Nanoseconds()

	return written, err
}
