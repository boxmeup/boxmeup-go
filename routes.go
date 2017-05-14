package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/cjsaylor/boxmeup-go/models"
	chain "github.com/justinas/alice"
)

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

func authHandler(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(res, "Authorization required.", 403)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(res, "Authorization header must be in the form of: Bearer {token}", 403)
			return
		}
		authConfig := models.AuthConfig{
			JWTSecret: config.JWTSecret,
		}
		claims, err := models.ValidateAndDecodeAuthClaim(parts[1], authConfig)
		if err != nil {
			http.Error(res, err.Error(), 403)
			return
		}
		newRequest := req.WithContext(context.WithValue(req.Context(), "user", claims))
		*req = *newRequest
		next.ServeHTTP(res, req)
		return
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
		chain.New(jsonResponseHandler).ThenFunc(LoginHandler),
	},
	Route{
		"CreateContainer",
		"POST",
		"/container",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(CreateContainerHandler),
	},
	Route{
		"Container",
		"GET",
		"/container/{id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(ContainerHandler),
	},
}
