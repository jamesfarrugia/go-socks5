package main

import (
	"net"
	"time"
)

// Functions for starting the service.  Somewhat useless since there is no client anymore
type serviceInit func(config) service

// Main loop for the chosen service (now always a 'server')
type serviceRunner func(service)

// Handler for authentication methods
type authHandler func([]byte, *serverConnection)

// Server configureation structure
type config struct {
	// true if it is a server, now always true
	server bool
	// listening host name
	host string
	// listening port
	port int
	// initialiser function
	init serviceInit
}

// Service structure
type service struct {
	// time when started
	started time.Time
	// true if it is running
	running bool
	// runner service function
	runner serviceRunner
	// service config
	config config
}

// server connection structure
type serverConnection struct {
	// ID of connection
	id int64
	// socket connection from client
	conn net.Conn
	// state of the connection
	status int
	// chosen authentication method ID (as per SOCKS5 protocol)
	authMethod int
	// chosen authentication method function
	auth authHandler
	// IP of target connection
	targetIP []byte
	// Port of target connection
	targetPort uint16
	// TCP connection to target
	targetConn net.Conn
	// Basic stats on number of bytes proxied
	dataCount int64
}
