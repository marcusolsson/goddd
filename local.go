package main

import (
	"net/http"

	"github.com/marcusolsson/goddd/server"
)

func main() {
	server.RegisterHandlers()

	http.ListenAndServe(":3000", nil)
}
