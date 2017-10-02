package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

var (
	re      *render.Render
	baseUrl *url.URL
)

func topHandler(w http.ResponseWriter, r *http.Request) {
	re.HTML(w, http.StatusOK, "index", "")
}

func getPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func postPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func main() {
	re = render.New(render.Options{
		Directory: "views",
	})

	r := mux.NewRouter()
	r.HandleFunc("/", topHandler)

	l := r.PathPrefix("/send").Subrouter()
	l.Methods("GET").HandlerFunc(getPushNotificationHandler)
	l.Methods("POST").HandlerFunc(postPushNotificationHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
