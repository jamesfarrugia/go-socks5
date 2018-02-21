package main

import "time"

// app meta struct
type Meta struct {
	// time when started
	Started time.Time
	// true if it is running
	Running bool
}
