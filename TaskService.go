package main

import (
	PeriodTask "assignment/api"
	"log"
	"net/http"
	"os"
)

func main() {
	// Get the listen address/port from the command-line argument
	listenAddr := "localhost:8080"
	if len(os.Args) > 1 {
		listenAddr = os.Args[1]
	}

	http.HandleFunc("/ptlist", PeriodTask.HandleGetRequest)

	log.Println("Server is listening on http://" + listenAddr)
	http.ListenAndServe(listenAddr, nil)
}

//go run main.go localhost:8080
