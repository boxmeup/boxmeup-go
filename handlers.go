package main

import (
	"encoding/json"
	"net/http"

	"strconv"

	"fmt"

	"github.com/cjsaylor/boxmeup-go/models"
	"github.com/gorilla/mux"
)

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
	result := make(chan models.UserResult)
	go models.GetUserById(db, id, result)
	user := <-result
	if user.Error != nil {
		res.WriteHeader(http.StatusNotFound)
		fmt.Fprint(res, "User not found")
	} else {
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(res).Encode(user.User); err != nil {
			panic(err)
		}
	}
}
