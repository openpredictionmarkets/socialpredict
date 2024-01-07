package handlers

import (
	"fmt"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Set the Content-Type header to indicate a JSON response
	w.Header().Set("Content-Type", "application/json")

	// Send a JSON-formatted response
	fmt.Fprint(w, `{"message": "Data From the Backend!"}`)

}
