package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func CreateNewTopic(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topicName")
	tokenString := r.Header.Get("token")
	token, ok := tokens.GetToken(tokenString)
	if !ok || !token.HasPermission(CreateTopic) {
		http.Error(w, "Token not found", http.StatusUnauthorized)
		return
	}

	_, err := GetOrCreateTopic(topicName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func PushEvent(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topicName")
	tokenString := r.Header.Get("token")
	token, ok := tokens.GetToken(tokenString)
	if !ok || !token.HasPermission(Write) || !token.IsSubscribed(topicName) {
		http.Error(w, "Token not found", http.StatusUnauthorized)
		return
	}
	body := r.Body
	defer body.Close()
	content := json.NewDecoder(body)
	eventContent := make(map[string]string)
	err := content.Decode(&eventContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	events := make([]Event, 0)
	for k, v := range eventContent {
		events = append(events, Event{
			Key:       k,
			Value:     v,
			Timestamp: time.Now(),
		})
	}
	topic, err := GetTopic(topicName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, event := range events {
		err = topic.LoadEvent(event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func PullEvent(w http.ResponseWriter, r *http.Request) {
	var seekNumber int64
	var err error
	topicName := r.PathValue("topicName")
	tokenString := r.Header.Get("token")
	token, ok := tokens.GetToken(tokenString)
	if !ok || !token.HasPermission(Read) || !token.IsSubscribed(topicName) {
		http.Error(w, "Token not found or does not have permission or is not subscribed", http.StatusUnauthorized)
		return
	}
	seek := r.URL.Query().Get("seek")
	if seek == "" {
		seekNumber = 0
	} else {
		seekNumber, err = strconv.ParseInt(seek, 0, 64)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	topic, err := GetTopic(topicName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	events, newSeek, err := topic.GetLastEvents(seekNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]any{
		"seek":   newSeek,
		"events": events,
	}
	json.NewEncoder(w).Encode(response)
}

func CreateToken(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("token")
	token, ok := tokens.GetToken(tokenString)
	if !ok || !token.HasPermission(CreateNewToken) {
		http.Error(w, "Token not found or does not have permission", http.StatusUnauthorized)
		return
	}
	bodyContent := make(map[string][]string)
	body := r.Body
	defer body.Close()
	content := json.NewDecoder(body)
	err := content.Decode(&bodyContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	permissions := bodyContent["permissions"]
	var permission Permission = 0
	for _, p := range permissions {
		switch p {
		case "read":
			permission |= Read
		case "write":
			permission |= Write
		case "createTopic":
			permission |= CreateTopic
		case "deleteTopic":
			permission |= DeleteTopic
		case "createToken":
			permission |= CreateNewToken
		case "deleteToken":
			permission |= DeleteToken
		case "subscribe":
			permission |= Subs
		case "admin":
			permission |= Admin
		}
	}
	var ts []string
	ts, ok = bodyContent["topics"]
	if !ok {
		ts = []string{}
	}
	newToken := NewToken(permission, ts...)
	err = json.NewEncoder(w).Encode(newToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func DeleteTokenEP(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("token")
	token, ok := tokens.GetToken(tokenString)
	if !ok || !token.HasPermission(DeleteToken) {
		http.Error(w, "Token not found or does not have permission", http.StatusUnauthorized)
		return
	}
	tokenToDelete := r.URL.Query().Get("token")
	if tokenToDelete == "" {
		http.Error(w, "No token to delete", http.StatusBadRequest)
	}
	tokens.RemoveToken(tokenToDelete)
	w.WriteHeader(http.StatusNoContent)
}

func Subscribe(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("token")
	token, ok := tokens.GetToken(tokenString)
	if !ok || !token.HasPermission(Subs) {
		http.Error(w, "Token not found or does not have permission", http.StatusUnauthorized)
		return
	}
	topicName := r.PathValue("topicName")
	token.Subscribe(topicName)
	w.WriteHeader(http.StatusCreated)
}

func GetCurrentToken(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("token")
	token, ok := tokens.GetToken(tokenString)
	if !ok {
		http.Error(w, "Token not found", http.StatusUnauthorized)
		return
	}
	err := json.NewEncoder(w).Encode(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
