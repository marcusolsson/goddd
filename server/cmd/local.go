package main

import (
	"net/http"

	"bitbucket.org/marcus_olsson/goddd/server"
)

func main() {
	server.RegisterHandlers()

	http.ListenAndServe(":3000", nil)
}
