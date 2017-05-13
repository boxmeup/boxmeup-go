package main

import "net/http"

// Route defines a route
type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.Handler
}

// Routes is a collection of routes
type Routes []Route

func jsonResponseHandler(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		next.ServeHTTP(res, req)
	}
	return http.HandlerFunc(fn)
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		http.HandlerFunc(IndexHandler),
	},
	Route{
		"Login",
		"POST",
		"/login",
		jsonResponseHandler(http.HandlerFunc(LoginHandler)),
	},
	Route{
		"User",
		"GET",
		"/user/{id}",
		jsonResponseHandler(http.HandlerFunc(UserHandler)),
	},
	Route{
		"CreateContainer",
		"POST",
		"/container",
		jsonResponseHandler(http.HandlerFunc(CreateContainerHandler)),
	},
}
