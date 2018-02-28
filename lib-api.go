package main

import "time"

// Meta app meta struct
type Meta struct {
	// time when started
	Started time.Time
	// true if it is running
	Running bool
}

// Config represents the service configuration
type Config struct {
	// listening host name
	Host string
	// listening port
	Port int
	// API host
	APIHost string
	// API port
	APIPort int
}

// Connection represents a connection to the server
type Connection struct {
	// time when started
	Started time.Time
	// ID of connection
	ID int64
	// state of the connection
	Status int
	// chosen authentication method ID (as per SOCKS5 protocol)
	AuthMethod int
	// IP of target connection
	TargetIP []byte
	// Port of target connection
	TargetPort uint16
	// Basic stats on number of bytes proxied
	DataIn int64
	// Basic stats on number of bytes proxied
	DataOut int64
	// ActiveTime is the total number of milliseconds the connection was active
	ActiveTime int64
}

// User represents an account that can be used to log in
type User struct {
	// Username
	Username string
	// Enabled flag
	Enabled bool
}
