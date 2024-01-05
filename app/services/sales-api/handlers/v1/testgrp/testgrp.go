package testgrp

import (
	"encoding/json"
	"net/http"
)

// Test is a basic HTTP Handler.
func Test(w http.ResponseWriter, r *http.Request) {
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	json.NewEncoder(w).Encode(status)
}
