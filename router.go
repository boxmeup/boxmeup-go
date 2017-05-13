package main

import (
	"github.com/gorilla/mux"
)

// NewRouter gets a pre-configured router with all defined routes
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.Handler)
	}
	return router
}
