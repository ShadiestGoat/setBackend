package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func RandInt(min, max int) int {
	rand.Seed(time.Now().UnixMilli())
	v := rand.Intn(max-min) + min
	return v
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var msgSucc = []byte(`{"status":"success"}`)

func StatusSuccess(w http.ResponseWriter) {
	Respond(w, 200, msgSucc)
}

// util function to write a unified error message
func WriteErr(w http.ResponseWriter, err *HTTPError) {
	fmt.Println(string(err.CachedMsg))
	Respond(w, err.Status, err.CachedMsg)
}

// util function for responding w/ a string
func RespondString(w http.ResponseWriter, status int, msg string) {
	Respond(w, status, []byte(msg))
}

// util function to respond w/ a status. Just puts the things in the same place
func Respond(w http.ResponseWriter, status int, msg []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(msg)
}

func CloseConn(conn *websocket.Conn) {
	conn.WriteControl(websocket.CloseMessage, []byte{}, time.Time{})
	conn.Close()
	conn = nil
}

func (c Card) MarshalJSON() ([]byte, error) {
	return []byte{
		'"',
		c[0],
		c[1],
		c[2],
		c[3],
		'"',
	}, nil
}
