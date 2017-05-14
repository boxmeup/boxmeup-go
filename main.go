package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	router := NewRouter()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Port), router))
}
