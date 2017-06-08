package main

import (
	"encoding/json"
	"log"
	"net/http"

	server "github.com/fesposito/go-ranger/http"
	"github.com/julienschmidt/httprouter"
)

func main() {
	s := server.NewHTTPServer()
	s.WithMiddleware(sampleMiddleware, anotherSampleMiddleware, server.RequestLog)
	s.WithDefaultErrorRoute()
	s.WithHealthCheckFor(nil)

	s.GET("/hello", helloEndpoint())

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

func helloEndpoint() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode("Hello world")
	}
}
