package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/cjsaylor/boxmeup-go/models/users"
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

type userKey string

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
		claims, err := users.ValidateAndDecodeAuthClaim(parts[1], users.AuthConfig{
			JWTSecret: config.JWTSecret,
		})
		if err != nil {
			http.Error(res, err.Error(), 403)
			return
		}
		var userKey userKey = "user"
		newRequest := req.WithContext(context.WithValue(req.Context(), userKey, claims))
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
		"Register",
		"POST",
		"/register",
		chain.New(jsonResponseHandler).ThenFunc(RegisterHandler),
	},
	Route{
		"CreateContainer",
		"POST",
		"/container",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(CreateContainerHandler),
	},
	Route{
		"UpdateContainer",
		"PUT",
		"/container/{id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(UpdateContainerHandler),
	},
	Route{
		"DeleteContainer",
		"DELETE",
		"/container/{id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(DeleteContainerHandler),
	},
	Route{
		"Container",
		"GET",
		"/container/{id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(ContainerHandler),
	},
	Route{
		"Containers",
		"GET",
		"/container",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(ContainersHandler),
	},
	Route{
		"CreateContainerItem",
		"POST",
		"/container/{id}/item",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(SaveContainerItemHandler),
	},
	Route{
		"ModifyContainerItem",
		"PUT",
		"/container/{id}/item/{item_id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(SaveContainerItemHandler),
	},
	Route{
		"Items",
		"GET",
		"/container/{id}/item",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(ContainerItemsHandler),
	},
	Route{
		"DeleteItems",
		"DELETE",
		"/container/{id}/item/{item_id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(DeleteContainerItemHandler),
	},
	Route{
		"ContainerQR",
		"GET",
		"/container/{id}/qrcode",
		http.HandlerFunc(ContainerQR),
	},
}
