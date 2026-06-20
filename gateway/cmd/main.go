package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok","service":"gateway"}`)
	})

	fmt.Println("🚀 Gateway starting on :8080")
	http.ListenAndServe(":8080", nil)
}
