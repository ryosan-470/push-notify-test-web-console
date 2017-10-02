package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	re.HTML(w, http.StatusOK, "push_notification_form", struct {
		Context context.Context
		Alert   bool
	}{
		r.Context(),
		false,
	})
}

func postPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	to := r.FormValue("to")
	title := r.FormValue("title")
	body := r.FormValue("body")
	alertType, message := sendPushNotificationViaHTTP(to, title, body)
	re.HTML(w, http.StatusOK, "push_notification_form",
		PushNotificationResponse{
			Context:   r.Context(),
			Alert:     true,
			AlertType: alertType,
			Message:   message,
		})
}

func sendPushNotificationViaHTTP(to, title, body string) (alertType, message string) {
	var payload = []byte(fmt.Sprintf(`{"to": "%s", "notification": {"title": "%s", "body": "%s"}}`, to, title, body))
	req, err := http.NewRequest("POST", FCM_ENDPOINT, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("key=%s", SERVER_KEY))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Printf("response Status: %d", resp.Status)
	log.Printf("response Headers: %s", resp.Header)
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var fcmHttpResponse FcmHttpResponse
	json.Unmarshal(responseBody, &fcmHttpResponse)

	if fcmHttpResponse.Failure != 0 {
		return "alert-danger", fmt.Sprintf("Failure to send notification to %s", to)
	}
	log.Printf("response Body: %s", string(responseBody))
	return "alert-success", fmt.Sprintf("Success to send notification to %s", to)
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
