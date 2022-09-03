package main

import "encoding/json"

type WSEvent string

const (
	E_SET WSEvent = "SET"

	// Only server -> client
	E_STFU WSEvent = "STFU"

	E_START WSEvent = "START"
	// Only server -> client
	E_JOINED WSEvent = "JOINED"
	E_LEFT   WSEvent = "LEFT"
	// Only server -> client
	E_ERR WSEvent = "ERROR"
)

type Event struct {
	Event WSEvent         `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type EventSET struct {
	PlayerID string   `json:"player"`
	Board    []*Card `json:"board"`
}
