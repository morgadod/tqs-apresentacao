package main

import (
	"net/http"
	"os"

	"tqs/internal/app"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("APP_PORT")
	// port := "8000"

	app, err := app.New()
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/survey", app.VoteSurvey).Methods("POST")

	http.ListenAndServe(":"+port, router)
}
