package main

import "context"

type PushNotificationResponse struct {
	Context   context.Context
	Alert     bool
	AlertType string
	Message   string
}

type FcmHttpResponse struct {
	MulticastId  int                      `json:multicast_id`
	Success      int                      `json:success`
	Failure      int                      `json:failure`
	CanonicalIds int                      `json:canonical_ids`
	Results      []FcmHttpResponseResults `json:results`
}

type FcmHttpResponseResults struct {
	MessageId      string `json:message_id`
	RegistrationId int    `json:registration_id`
	Error          string `json:error`
}
