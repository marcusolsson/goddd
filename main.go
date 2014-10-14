package main

import (
	"net/http"
	"os"

	"github.com/marcusolsson/goddd/api"
)

func main() {
	api.RegisterHandlers()

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
