package api

import (
	"encoding/json"
	"net/http"
)

// test function
func testHistory(w http.ResponseWriter, r *http.Request) {

	mockData := map[string]string{"data": "here is some mock data"}

	marshalled, err := json.Marshal(mockData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(marshalled)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
