package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func topHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World\n"))
}

func sendPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func listPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", topHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
