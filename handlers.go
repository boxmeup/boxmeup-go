package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cjsaylor/boxmeup-go/models"
	"github.com/gorilla/mux"
)

func Index(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	res.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res.WriteHeader(http.StatusOK)
	user := models.User{
		ID:            10,
		Email:         vars["email"],
		Password:      "blah",
		IsActive:      true,
		ResetPassword: false,
		Created:       time.Now(),
		Modified:      time.Now(),
	}
	if err := json.NewEncoder(res).Encode(user.ToSafeUser()); err != nil {
		panic(err)
	}
}
