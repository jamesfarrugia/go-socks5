package main

import (
	"encoding/json"
	"fmt"
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
	// connections
	// users
	// blacklists

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
