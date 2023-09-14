package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/flow"
	"github.com/guilhermeabel/discord-bot/response"
)

type applicationServer struct{}

func (app *applicationServer) server() *http.Server {
	return &http.Server{
		Addr:         ":80",
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

func (app *applicationServer) routes() http.Handler {
	mux := flow.New()

	mux.Use(app.recoverPanic)

	mux.HandleFunc("/", hello, http.MethodGet)

	return mux
}

func (app *applicationServer) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			err := recover()
			if err != nil {
				serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func serverError(w http.ResponseWriter, err error) {
	message := "The server encountered a problem and could not process your request"
	var data []string
	response.Error(w, http.StatusInternalServerError, message, data)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<h3>bot running</h3>
	`))
}
