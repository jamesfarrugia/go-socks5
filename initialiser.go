package main

import (
	"strconv"
	"strings"
)

func doInit(args []string) config {
	conf := config{server: false, host: "127.0.0.1", port: 1080}

	for _, arg := range args {
		conf.server = true
		conf.init = doStartServer

		if strings.HasPrefix(arg, "-host=") {
			conf.host = arg[6:]
		}

		if strings.HasPrefix(arg, "-port=") {
			portStr := string(arg[6:])
			var err error = nil
			conf.port, err = strconv.Atoi(portStr)
			if err != nil {
				panic("Port must be a number")
			}
		}
	}

	return conf
}
