package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	router := NewRouter()
	env := EnvConfig()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", env.Port), router))
}
