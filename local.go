package main

import (
	"net/http"

	"github.com/marcusolsson/goddd/api"
)

func main() {
	api.RegisterHandlers()

	http.ListenAndServe(":3000", nil)
}
