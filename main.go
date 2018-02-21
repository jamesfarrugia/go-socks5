package main

import (
	"os"
	"time"

	logging "github.com/op/go-logging"
)

// Server state
var app appState

// Main function
func main() {
	logging.SetBackend(backend, backendFormatter)
	log.Info("Golang SOCKS5 app - James Farrugia 2018")
	conf := doInit(os.Args[1:])
	svc := conf.init(conf)

	go func() {
		err := doStartAPI(conf.apiHost, conf.apiPort)
		if err != nil {
			log.Error(err.Error())
		}
	}()

	log.Info("Starting service at ", svc.started)
	svc.runner(svc)
	log.Info("Stopped service at ", time.Now())
}

// Starts the proxy server
func doStartServer(conf config) service {
	log.Info("Starting SOCKS5 server")
	log.Info("Shall listen on", conf.host, conf.port)
	return service{started: time.Now(), running: true, runner: server, config: conf}
}
