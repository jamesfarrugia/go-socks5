package main

import (
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("socks5-app")

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ \t%{level:.4s} %{id:03x}%{color:reset} %{message}",
)

var backend = logging.NewLogBackend(os.Stderr, "", 0)
var backendFormatter = logging.NewBackendFormatter(backend, format)