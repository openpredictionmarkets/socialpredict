package helpers

import (
	"errors"
	"log"
	"net/http"

	"gorm.io/gorm"
)

func HandleError(w http.ResponseWriter, err error, message string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "User not found", http.StatusNotFound)
	} else {
		http.Error(w, message, http.StatusInternalServerError)
		log.Printf("%s: %v", message, err)
	}
}
