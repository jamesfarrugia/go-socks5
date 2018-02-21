package main

import (
	"os"
	"time"
)

func main() {
	log.Info("Golang SOCKS5 app - James Farrugia 2018")
	conf := doInit(os.Args[1:])
	service := conf.init(conf)

	log.Info("Starting service at ", service.started)
	service.runner(service)
	log.Info("Stopped service at ", time.Now())
}

func doStartServer(conf config) service {
	log.Info("Starting SOCKS5 server")
	log.Info("Shall listen on", conf.host, conf.port)
	return service{started: time.Now(), running: true, runner: server, config: conf}
}
