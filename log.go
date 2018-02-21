package main

import (
	"os"

	"github.com/op/go-logging"
)

// Logger
var log = logging.MustGetLogger("socks5-app")

// Formatted with timestamp and level
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ \t%{level:.4s} %{id:03x}%{color:reset} %{message}",
)

// Logger backend
var backend = logging.NewLogBackend(os.Stderr, "", 0)

// Backend formatter
var backendFormatter = logging.NewBackendFormatter(backend, format)
