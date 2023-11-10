package server

import (
	"fmt"
	"log"
	"net/http"
)

func Start() {
	http.HandleFunc("/", helloHandler)
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, worlds!!!")
}
