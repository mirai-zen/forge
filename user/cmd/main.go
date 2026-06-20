package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok","service":"user"}`)
	})

	fmt.Println("🚀 User Service starting on :8081")
	http.ListenAndServe(":8081", nil)
}
