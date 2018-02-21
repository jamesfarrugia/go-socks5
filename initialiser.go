package main

import (
	"strconv"
	"strings"
)

// Goes through cmd args and sets up the host, port and service type
func doInit(args []string) config {
	conf := config{server: true, host: "127.0.0.1", port: 1080}
	conf.server = true
	conf.init = doStartServer

	for _, arg := range args {
		if strings.HasPrefix(arg, "-host=") {
			conf.host = arg[6:]
		}

		if strings.HasPrefix(arg, "-port=") {
			portStr := arg[6:]
			var err error
			conf.port, err = strconv.Atoi(portStr)
			if err != nil {
				panic("Port must be a number")
			}
		}
	}

	return conf
}
