package main

import (
	"log"
	"net/http"

	server "github.com/fesposito/go-ranger/http"
)

func main() {
	s := server.NewHTTPServer()
	s.WithDefaultErrorRoute()
	s.WithHealthCheckFor(nil)
	//s.WithMiddleware(SampleMiddleware)
	s.Start(":8080")
}

func SampleMiddleware(h http.Handler) http.Handler {
	log.Print("middleware!")
	return h
}
