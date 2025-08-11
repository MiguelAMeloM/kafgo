package main

import (
	"time"
)

type Event struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type EventsList []Event
