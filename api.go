package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func doStartAPI(host string, port int) error {
	log.Info("Starting HTTP service at", host, ":", port, "...")
	htinfo := fmt.Sprintf("%s:%d", host, port)

	log.Info("Preparing API")
	router := doInitAPI()

	log.Info("Serving")
	err := http.ListenAndServe(htinfo, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		return err
	}

	return nil
}

func doInitAPI() *httprouter.Router {
	router := httprouter.New()

	// meta
	router.GET("/", httpInfo)

	// status
	router.GET("/status", httpStatus)

	// config
	router.GET("/config", httpConfig)

	// connections
	router.GET("/connections", httpConnections)

	// users
	router.GET("/users", httpUsers)

	// blacklists

	router.GET("/reverse/:ip", httpReverseIP)
	return router
}

func httpInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	meta, err := json.Marshal(Meta{Started: app.service.started, Running: app.service.running})

	if err != nil {
		log.Error("[API] - httpInfo - ", err)
	}

	fmt.Fprintf(w, "%s", meta)
}

func httpStatus(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	meta, err := json.Marshal(len(app.connections))

	if err != nil {
		log.Error("[API] - httpStatus - ", err)
	}

	fmt.Fprintf(w, "%s", meta)
}

func httpConfig(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	conf, err := json.Marshal(Config{
		Host:    app.service.config.host,
		Port:    app.service.config.port,
		APIHost: app.service.config.apiHost,
		APIPort: app.service.config.apiPort})

	if err != nil {
		log.Error("[API] - httpStatus - ", err)
	}

	fmt.Fprintf(w, "%s", conf)
}

func httpConnections(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	var connections = make([]Connection, 1)
	app.connLock.Lock()
	for _, con := range app.connections {
		connections = append(connections, Connection{
			ID:         con.id,
			Started:    con.started,
			Status:     con.status,
			AuthMethod: con.authMethod,
			TargetIP:   con.targetIP,
			TargetPort: con.targetPort,
			DataIn:     con.dataIn,
			DataOut:    con.dataOut,
			ActiveTime: con.activeTime})
	}
	app.connLock.Unlock()

	connsJSON, err := json.Marshal(connections)

	if err != nil {
		log.Error("[API] - httpStatus - ", err)
	}

	fmt.Fprintf(w, "%s", connsJSON)
}

func httpUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	var users []User
	app.userLock.Lock()
	users = append(users, User{
		Username: "james",
		Enabled:  true})
	app.userLock.Unlock()

	usersJSON, err := json.Marshal(users)

	if err != nil {
		log.Error("[API] - httpUsers - ", err)
	}

	fmt.Fprintf(w, "%s", usersJSON)
}

func httpReverseIP(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	ip := p.ByName("ip")
	addr, err := net.LookupAddr(ip)

	if err != nil {
		log.Error("[API] - httpReverseIP - ", err)
	}

	addressesJSON, err := json.Marshal(addr)
	if err != nil {
		log.Error("[API] - httpReverseIP - ", err)
	}

	fmt.Fprintf(w, "%s", addressesJSON)
}
