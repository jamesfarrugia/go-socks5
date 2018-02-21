package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

func doProxy(conn *serverConnection) {
	conn.status = stsRdConn
	doProcessesConnection(conn)
	doConnectToTarget(conn)

	v := make([]byte, 10)
	v[0] = 0x05 // SOCKS 5
	v[1] = 0x00 // STATUS, default to success (0), set to 0x1 if we had tgt err
	v[2] = 0x00 // always 0
	v[3] = 0x01 // IP4 type

	v[4] = conn.targetIp[0] // IP4 Octet 1
	v[5] = conn.targetIp[1] // IP4 Octet 2
	v[6] = conn.targetIp[2] // IP4 Octet 3
	v[7] = conn.targetIp[3] // IP4 Octet 4

	v[8] = 0x0 // IP4 Port
	v[9] = 0x0 // IP4 Port

	if conn.status == stsTgtErr {
		v[1] = 0x1
		conn.status = stsClose
		conn.conn.Write(v)
		return
	}

	conn.conn.Write(v)

	// process connection
	doHandleProxying(conn)
}

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
				if cmdBuf[0] != 5 {
					log.Error("Version mismatch")
					conn.status = stsClose
					continue
				}
				if cmdBuf[1] != 1 {
					log.Error("Only TCP/IP proxying is supported")
					conn.status = stsClose
					continue
				}
				if cmdBuf[3] != 1 {
					log.Error("Only IP4 addresses are supported, want 1 got", cmdBuf[3])
					conn.status = stsClose
					continue
				}

				if len(cmdBuf) == 10 {
					ip := cmdBuf[4:8]
					port := cmdBuf[8:]
					conn.status = stsTgtConn
					conn.targetIp = ip
					conn.targetPort = binary.BigEndian.Uint16(port)
				}

			}
		}
	}
}

func doConnectToTarget(conn *serverConnection) {
	ip := net.IPv4(conn.targetIp[0], conn.targetIp[1], conn.targetIp[2], conn.targetIp[3])
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
		pipe(conn, conn.conn, conn.targetConn)
	}
}

// ingress from target to client
func ingress(conn *serverConnection) {
	for conn.status == stsProxying {
		pipe(conn, conn.targetConn, conn.conn)
	}
}

func pipe(conn *serverConnection, from net.Conn, to net.Conn) {
	buf := make([]byte, 1)
	for conn.status == stsProxying {
		_, err := from.Read(buf)
		if err != nil {
			if err == io.EOF {
				conn.status = stsClose
			} else if err != nil {
				log.Error("Failed to read", err.Error())
				conn.status = stsClose
			}
			break
		}
		to.Write(buf)
	}
}
