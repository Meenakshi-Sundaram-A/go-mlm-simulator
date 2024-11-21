// main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func processData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check the plan type to determine which processing function to call
	planType := data["plan_type"].(string)
	var results []map[string]interface{}
	if planType == "binary" {
		results = ProcessBinaryTree(data)
	} else if planType == "unilevel" {
		fmt.Println("Yes Unilevel")
		results = ProcessUnilevelTree(data)
	} else {
		http.Error(w, "Invalid plan type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/api/processData", processData)
	fmt.Println("Go server is listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
