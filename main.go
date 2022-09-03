package main

import "net/http"

func main() {
	InitConfig()
	InitDB()

	http.ListenAndServe(":5122", routerAPI())
}
