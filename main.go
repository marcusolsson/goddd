package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/marcusolsson/goddd/api"
)

var port int

func main() {
	flag.IntVar(&port, "port", 8080, "the server port")
	flag.Parse()

	api.RegisterHandlers()

	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
