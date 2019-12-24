package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/scaleUp", ScaleUp)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}