package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/unrolled/render"
)

var (
	re      *render.Render
	baseUrl *url.URL
	db      *sql.DB
)

const (
	DB_FILENAME  = "db.sqlite3"
	FCM_ENDPOINT = "https://fcm.googleapis.com/fcm/send"
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

	log.Printf("response Status: %s", resp.Status)
	log.Printf("response Headers: %s", resp.Header)
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error occured when decoding response body")
		panic(err)
	}
	var fcmHTTPResponse FcmHttpResponse
	json.Unmarshal(responseBody, &fcmHTTPResponse)

	if err = insertResultToDB(to, title, body, fcmHTTPResponse); err != nil {
		log.Print(err)
	}

	if fcmHTTPResponse.Failure != 0 {
		return "alert-danger", fmt.Sprintf("Failure to send notification to %s", to)
	}
	log.Printf("response Body: %s", string(responseBody))
	return "alert-success", fmt.Sprintf("Success to send notification to %s", to)
}

func insertResultToDB(to, title, body string, response FcmHttpResponse) error {
	isSuccess := true
	if response.Failure != 0 {
		isSuccess = false
	}

	stmt, err := db.Prepare(`INSERT INTO notification_list("to", "title", "body", "is_success", "date") VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("INSERTING ERROR: %v", err)
	}

	_, err = stmt.Exec(to, title, body, isSuccess, fmt.Sprintf("%s", time.Now().Format("2006/01/02 15:04:05 MST")))
	if err != nil {
		return fmt.Errorf("INSERTING ERROR: %v", err)
	}
	return nil
}

func initialize() {
	log.Printf("initialize()")
	_, err := os.Create(DB_FILENAME)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", DB_FILENAME, err)
	}
	db, err = sql.Open("sqlite3", DB_FILENAME)
	if err != nil {
		log.Fatalf("Failed to connect the DB: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE notification_list ("id" integer PRIMARY KEY AUTOINCREMENT, "to" text, "title" text, "body" text, "is_success" integer, "date" text)`)
	if err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}
}

func main() {
	_, err := os.Stat(DB_FILENAME)
	if err != nil {
		initialize()
	}

	db, err = sql.Open("sqlite3", DB_FILENAME)
	if err != nil {
		log.Fatalf("Failed to connect the DB: %s", err.Error())
	}
	defer db.Close()

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
