package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// NOOP Authentication method
const autMtNone = 0

// Username/PAssword authentication method
const autMtPass = 2

// Connection status
const (
	// New connection
	stsNew = iota
	// Reading header
	stsRdHead = iota
	// Negotiating auth
	stsNegAuth = iota
	// No auth methods error
	stsErNoAuth = iota
	// Write auth response
	stsWrCmd = iota
	// Write auth error
	stsWrAuthErr = iota
	// Read command
	stsRdConn = iota
	// Open target connection
	stsTgtConn = iota
	// Target error
	stsTgtErr = iota
	// Proxying connection
	stsProxying = iota
	// Closing connection
	stsClose = iota
)

// Main entry point for the SOCKS5 server
// This will open a passive TCP port, accept clients
// and process them in their own routine
func server(serv service) {
	log.Info("Starting server")
	app.service = &serv
	app.connLock = &sync.Mutex{}
	app.userLock = &sync.Mutex{}

	addr := fmt.Sprintf("%s:%d", serv.config.host, serv.config.port)
	server, err := net.Listen("tcp", addr)

	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %s", err.Error()))
	}

	go garbageCollector()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Error("Failed to accept connection: ", err.Error())
		}

		go doHandleConnection(conn)
	}
}

// Main entry point for a new connection
// The authentication is negotiated, processed, and if there were no errors, proxying starts
func doHandleConnection(conn net.Conn) {
	log.Debug("Accepted connection", conn.RemoteAddr().String())

	now := time.Now()
	sc := serverConnection{
		id:         now.UnixNano(),
		conn:       conn,
		status:     stsNew,
		authMethod: -1,
		started:    now}

	app.connLock.Lock()
	app.connections = append(app.connections, &sc)
	app.connLock.Unlock()

	if sc.status == stsNew {
		sc.status = stsRdHead
	}

	doProcessAuth(&sc)

	for sc.status == stsWrCmd {

		// write auth success
		v := make([]byte, 1)
		v[0] = 0x00
		_, err := sc.conn.Write(v)
		if err != nil {
			log.Error("Failed to write auth success response")
		}

		doProxy(&sc)
	}

	log.Info("Closing connection", conn.RemoteAddr().String(), "status", sc.status)
	err := conn.Close()
	if err != nil {
		log.Error("Failed to close connection")
	}
}

// Process the connection authentication
func doProcessAuth(sc *serverConnection) {

	doNegotation(sc)
	doAuthentication(sc)
}

// Negotiates the connection
func doNegotation(sc *serverConnection) {
	/*	1: SOCKS version (must be 0x5)
		2: No. of auth methods (0x2, noop and user/pass only supported)
		3: Auth method (0x0 noop, 0x2 for user/pass) */
	for sc.status == stsRdHead {
		headBuf := make([]byte, 0, 256)
		tmp := make([]byte, 32)
		for sc.status == stsRdHead {
			n, err := sc.conn.Read(tmp)
			if err != nil {
				if err == io.EOF {
					sc.status = stsClose
				} else if err != nil {
					log.Error("Failed to read", err.Error())
				}
				break
			}
			headBuf = append(headBuf, tmp[:n]...)
			doProcessHeader(headBuf, sc)
		}
	}

	doNegotationResponse(sc)
}

// Processes a negotiation header
func doProcessHeader(data []byte, conn *serverConnection) {

	if len(data) == 0 {
		return
	}

	if data[0] != 0x05 {
		conn.status = stsClose
		return
	}

	if len(data) == 1 {
		return
	}

	log.Info("Client supports ", data[1], " auth methods")

	if len(data) >= (int(data[1]) + 2) {
		// will choose the highest value automatically, between none and user/pass
		for _, code := range data[2:] {
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
			log.Info("No valid auth methods")
			conn.status = stsErNoAuth
		} else {
			conn.status = stsNegAuth
		}
	}
}

// Sends the response for the negotiation.  Sends the SOCKS5 version (0x5), and an error if that is the case.
// The chosen repsonse is sent in the doAuthentication function, which has a more direct inteaction with the
//method itself
func doNegotationResponse(sc *serverConnection) {
	if sc.status == stsErNoAuth || sc.status == stsNegAuth {
		// write socks version
		v := make([]byte, 1)
		v[0] = 0x05
		_, err := sc.conn.Write(v)
		if err != nil {
			log.Error("Failed to write auth response")
		}
	}

	if sc.status == stsErNoAuth {
		// write error
		v := make([]byte, 1)
		v[0] = 0xff
		_, err := sc.conn.Write(v)
		if err != nil {
			log.Error("Failed to write no auth method response")
		}
		sc.status = stsClose
	}
}

// Sends the chosen method and expects the client to communicate accordingly.
// The data is passed to the relevant authentication function to handle the
// authentication state
func doAuthentication(sc *serverConnection) {
	for sc.status == stsNegAuth {
		v := make([]byte, 1)
		v[0] = byte(sc.authMethod)
		_, err := sc.conn.Write(v)
		if err != nil {
			log.Error("Failed to write auth method response")
		}

		negBuf := make([]byte, 0, 516)
		tmp := make([]byte, 32)
		for sc.status == stsNegAuth {
			n, err := sc.conn.Read(tmp)
			if err != nil {
				if err == io.EOF {
					sc.status = stsClose
				} else if err != nil {
					log.Error("Failed to read", err.Error())
				}
				break
			}
			negBuf = append(negBuf, tmp[:n]...)
			sc.auth(negBuf, sc)
		}
	}

	doAuthenticationResponse(sc)
}

// Reponds to the connection info according to the status set by the
// authentication function
func doAuthenticationResponse(sc *serverConnection) {
	if sc.status == stsWrAuthErr || sc.status == stsWrCmd {
		// write auth version
		v := make([]byte, 1)
		v[0] = 0x01
		_, err := sc.conn.Write(v)
		if err != nil {
			log.Error("Failed to write auth response byte 0")
		}
	}

	if sc.status == stsWrAuthErr {
		// write auth error
		v := make([]byte, 1)
		v[0] = 0x01
		_, err := sc.conn.Write(v)
		if err != nil {
			log.Error("Failed to write auth error response")
		}
		// close connection
		sc.status = stsClose
	}
}

// Filter the active connections to only keep those that are not closed
func garbageCollector() {
	for app.service.running {
		var filtered []*serverConnection

		app.connLock.Lock()
		for _, conn := range app.connections {
			if conn.status != stsClose {
				filtered = append(filtered, conn)
			}
		}

		app.connections = filtered
		app.connLock.Unlock()
		time.Sleep(time.Duration(5) * time.Second)
	}
}
