package main

import (
	"log"
	"net/http"

	server "github.com/fesposito/go-ranger/http"
)

func main() {
	s := server.NewHTTPServer()
	s.WithMiddleware(sampleMiddleware, anotherSampleMiddleware, server.RequestLog)
	s.WithDefaultErrorRoute()
	s.WithHealthCheckFor(nil)
	s.Start(":8080")
}

func sampleMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Print("middleware!")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func anotherSampleMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Print("another middleware!")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
