package routing

import (
	"context"
	"net/http"
	"strings"

	"github.com/cjsaylor/boxmeup-go/modules/config"
	"github.com/cjsaylor/boxmeup-go/modules/users"
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
			http.Error(res, "Authorization required.", 401)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(res, "Authorization header must be in the form of: Bearer {token}", 401)
			return
		}
		claims, err := users.ValidateAndDecodeAuthClaim(parts[1], users.AuthConfig{
			JWTSecret: config.Config.JWTSecret,
		})
		if err != nil {
			http.Error(res, err.Error(), 401)
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
		"/api/user/login",
		chain.New(jsonResponseHandler).ThenFunc(LoginHandler),
	},
	Route{
		"Register",
		"POST",
		"/api/user/register",
		chain.New(jsonResponseHandler).ThenFunc(RegisterHandler),
	},
	Route{
		"CreateContainer",
		"POST",
		"/api/container",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(CreateContainerHandler),
	},
	Route{
		"UpdateContainer",
		"PUT",
		"/api/container/{id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(UpdateContainerHandler),
	},
	Route{
		"DeleteContainer",
		"DELETE",
		"/api/container/{id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(DeleteContainerHandler),
	},
	Route{
		"Containers",
		"GET",
		"/api/container",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(ContainersHandler),
	},
	Route{
		"ContainerQR",
		"GET",
		"/api/container/{id}/qrcode",
		chain.New(authHandler).ThenFunc(ContainerQR),
	},
	Route{
		"CreateContainerItem",
		"POST",
		"/api/container/{id}/item",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(SaveContainerItemHandler),
	},
	Route{
		"ModifyContainerItem",
		"PUT",
		"/api/container/{id}/item/{item_id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(SaveContainerItemHandler),
	},
	Route{
		"DeleteItems",
		"DELETE",
		"/api/container/{id}/item/{item_id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(DeleteContainerItemHandler),
	},
	Route{
		"Items",
		"GET",
		"/api/container/{id}/item",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(ContainerItemsHandler),
	},
	Route{
		"CreateLocation",
		"POST",
		"/api/location",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(CreateLocationHandler),
	},
	Route{
		"UpdateLocation",
		"PUT",
		"/api/location/{id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(UpdateLocationHandler),
	},
	Route{
		"DeleteLocation",
		"DELETE",
		"/api/location/{id}",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(DeleteLocationHandler),
	},
	Route{
		"Locations",
		"GET",
		"/api/location",
		chain.New(authHandler, jsonResponseHandler).ThenFunc(LocationsHandler),
	},
}
