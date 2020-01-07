package main

import (
	"net/http"
	fairbinden "slack_fairbinden"
)

func main() {
	http.HandleFunc("/fairbinden", fairbinden.SendSlack)
	http.ListenAndServe(":18082", nil)
}
