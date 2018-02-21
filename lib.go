package main

import (
	"net"
	"time"
)

type serviceInit func(config) service
type serviceRunner func(service)
type authHandler func([]byte, *serverConnection)

type config struct {
	server bool
	host   string
	port   int
	init   serviceInit
}

type service struct {
	started time.Time
	running bool
	runner  serviceRunner
	config  config
}

type serverConnection struct {
	conn       net.Conn
	status     int
	authMethod int
	auth       authHandler
	targetIp   []byte
	targetPort uint16
	targetConn net.Conn
	dataCount  int8
}
