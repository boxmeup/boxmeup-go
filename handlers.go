package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"strconv"

	"github.com/cjsaylor/boxmeup-go/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type jsonErrorResponse struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

// IndexHandler serves the static page
func IndexHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Welcome!")
}

// LoginHandler authenticates via email and password
func LoginHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()

	userModel := models.UserStore{DB: db}
	token, err := userModel.Login(
		models.AuthConfig{
			LegacySalt: config.LegacySalt,
			JWTSecret:  config.JWTSecret,
		},
		req.PostFormValue("email"),
		req.PostFormValue("password"))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
		jsonOut.Encode(jsonErrorResponse{-1, "Authentication failure."})
	} else {
		res.WriteHeader(http.StatusOK)
		jsonOut.Encode(map[string]string{
			"token": token,
		})
	}
}

// CreateContainerHandler allows creation of a container from a POST method
// Expected body:
//   name
func CreateContainerHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value("user").(jwt.MapClaims)["id"].(float64))
	userModel := models.UserStore{DB: db}
	user, err := userModel.ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "User specified not found."})
		return
	}
	container := models.Container{
		User: user,
		Name: req.PostFormValue("name"),
	}
	containerModel := models.ContainerStore{DB: db}
	err = containerModel.Create(&container)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(jsonErrorResponse{-2, "Failed to create the container."})
	} else {
		res.WriteHeader(http.StatusOK)
		jsonOut.Encode(map[string]int64{
			"id": container.ID,
		})
	}
}

// ContainerHandler gets a specific container by ID
func ContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value("user").(jwt.MapClaims)["id"].(float64))
	containerID, _ := strconv.Atoi(vars["id"])
	containerModel := models.ContainerStore{DB: db}
	container, err := containerModel.ByID(int64(containerID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "Container not found."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(jsonErrorResponse{-2, "Not allowed to view this container."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(container)
}

// ContainersHandler gets all user containers
func ContainersHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value("user").(jwt.MapClaims)["id"].(float64))
	userModel := models.UserStore{DB: db}
	user, err := userModel.ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(jsonErrorResponse{-1, "Unable to get user information."})
		return
	}
	params := req.URL.Query()
	var limit models.QueryLimit
	page, _ := strconv.Atoi(params.Get("page"))
	limit.SetPage(page, models.ContainerQueryLimit)
	containerModel := models.ContainerStore{DB: db}
	sort := containerModel.GetSortBy(params.Get("sort_field"), models.SortType(params.Get("sort_dir")))
	response, err := containerModel.UserContainers(user, sort, limit)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(jsonErrorResponse{-2, "Unable to retrieve containers."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(response)
}
