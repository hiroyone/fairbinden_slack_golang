package main

import (
	fairbinden "fairbindenlunch"
	"net/http"
)

func main() {
	http.HandleFunc("/fairbinden", fairbinden.SendSlack)
	http.ListenAndServe(":18082", nil)
}
