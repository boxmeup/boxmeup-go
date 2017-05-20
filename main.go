package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cjsaylor/boxmeup-go/modules/config"
	"github.com/cjsaylor/boxmeup-go/modules/routing"
)

func main() {
	router := routing.NewRouter()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.Port), router))
}
