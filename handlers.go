package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cjsaylor/boxmeup-go/models"
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

// UserHandler is a route handler for getting specific user information
func UserHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, err := GetDBResource()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		panic(err)
	}
	user, err := models.GetUserByID(db, id)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		json.NewEncoder(res).Encode(jsonErrorResponse{-1, "User not found."})
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(user)
	}
}

// CreateContainerHandler allows creation of a container from a POST method
func CreateContainerHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()
	defer db.Close()
	userID, _ := strconv.Atoi(req.PostFormValue("user_id"))
	user, err := models.GetUserByID(db, userID)
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
	err = models.CreateContainer(db, &container)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(jsonErrorResponse{-2, "Failed to create the container."})
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(struct {
			ID int64 `json:"id"`
		}{
			ID: container.ID,
		})
	}
}
