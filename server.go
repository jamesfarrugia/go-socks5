package main

import (
	"fmt"
	"io"
	"net"
)

const autMtNone = 0
const autMtPass = 2

const (
	stsNew       = iota
	stsRdHead    = iota
	stsNegAuth   = iota
	stsErNoAuth  = iota
	stsWrCmd     = iota
	stsWrAuthErr = iota
	stsRdConn    = iota
	stsTgtConn   = iota
	stsTgtErr    = iota
	stsProxying  = iota
	stsClose     = iota
)

func server(serv service) {
	log.Info("Starting server")

	addr := fmt.Sprintf("%s:%d", serv.config.host, serv.config.port)
	server, err := net.Listen("tcp", addr)

	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %s", err.Error()))
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			panic(fmt.Sprintf("Failed to accept connection: %s", err.Error()))
		}

		go doHandleConnection(conn)
	}
}

func doHandleConnection(conn net.Conn) {
	log.Info("Accepted connection", conn.RemoteAddr().String())

	sc := serverConnection{conn: conn, status: stsNew, authMethod: -1}

	for sc.status == stsNew {
		sc.status = stsRdHead
	}

	log.Debug("New connection")

	/* 1: SOCKS version (must be 0x5)
	   2: No. of auth methods (0x1, user/pass only supported)
	3: Auth method (0x2 for user/pass) */
	for sc.status == stsRdHead {
		log.Debug("read header")
		headBuf := make([]byte, 0, 256)
		tmp := make([]byte, 32)
		for sc.status == stsRdHead {
			log.Debug("wait read", sc.status)
			n, err := conn.Read(tmp)
			if err != nil {
				if err == io.EOF {
					sc.status = stsClose
				} else if err != nil {
					log.Error("Failed to read", err.Error())
				}
				break
			}
			log.Debug("read bytes")
			headBuf = append(headBuf, tmp[:n]...)
			doProcessHeader(headBuf, &sc)
		}
	}

	log.Debug("header done ", sc.status)

	if sc.status == stsErNoAuth || sc.status == stsNegAuth {
		// write socks version
		v := make([]byte, 1)
		v[0] = 0x05
		sc.conn.Write(v)
	}

	if sc.status == stsErNoAuth {
		// write error
		v := make([]byte, 1)
		v[0] = 0xff
		sc.conn.Write(v)
		// close connection
		sc.status = stsClose
	}

	for sc.status == stsNegAuth {
		log.Debug("chosen method =", sc.authMethod)
		v := make([]byte, 1)
		v[0] = byte(sc.authMethod)
		sc.conn.Write(v)

		log.Debug("read auth")
		negBuf := make([]byte, 0, 516)
		tmp := make([]byte, 32)
		for sc.status == stsNegAuth {
			n, err := conn.Read(tmp)
			if err != nil {
				if err == io.EOF {
					sc.status = stsClose
				} else if err != nil {
					log.Error("Failed to read", err.Error())
				}
				break
			}
			negBuf = append(negBuf, tmp[:n]...)
			sc.auth(negBuf, &sc)
		}
	}

	log.Debug("header done ", sc.status)

	if sc.status == stsWrAuthErr || sc.status == stsWrCmd {
		// write auth version
		v := make([]byte, 1)
		v[0] = 0x01
		sc.conn.Write(v)
	}

	if sc.status == stsWrAuthErr {
		// write auth error
		v := make([]byte, 1)
		v[0] = 0x01
		sc.conn.Write(v)
		// close connection
		sc.status = stsClose
	}

	for sc.status == stsWrCmd {

		// write auth success
		v := make([]byte, 1)
		v[0] = 0x00
		sc.conn.Write(v)

		log.Debug("process command")
		doProxy(&sc)
	}

	log.Info("Closing connection", conn.RemoteAddr().String(), "status", sc.status)
	conn.Close()
}

func doProcessHeader(data []byte, conn *serverConnection) {

	if len(data) == 0 {
		return
	}

	if data[0] != 0x05 {
		conn.status = stsClose
		return
	}

	log.Debug("SOCKS version 5")

	if len(data) == 1 {
		return
	}

	log.Debug("Client supports ", data[1], " auth methods")

	if len(data) >= (int(data[1]) + 2) {
		log.Debug("processing auth methods")
		// will choose the highest value automatically, between none and user/pass
		for _, code := range data[2:] {
			log.Debug("AUTH METHOD", code)
			if code == autMtNone {
				conn.authMethod = autMtNone
				conn.auth = doAuthNone
			}

			if code == autMtPass {
				conn.authMethod = autMtPass
				conn.auth = doAuthPass
			}
		}

		if conn.authMethod == -1 {
			log.Debug("No valid auth methods")
			conn.status = stsErNoAuth
		} else {
			conn.status = stsNegAuth
		}
	}
}
