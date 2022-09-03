package main

import (
	"math/rand"
	"net/http"
	"time"
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
