package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"strconv"

	"github.com/cjsaylor/boxmeup-go/models"
	"github.com/cjsaylor/boxmeup-go/modules/containers"
	"github.com/cjsaylor/boxmeup-go/modules/items"
	"github.com/cjsaylor/boxmeup-go/modules/locations"
	"github.com/cjsaylor/boxmeup-go/modules/users"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	qrcode "github.com/skip2/go-qrcode"
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

	token, err := users.NewStore(db).Login(
		users.AuthConfig{
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

// RegisterHandler creates new users.
func RegisterHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()

	email := req.PostFormValue("email")
	password := req.PostFormValue("password")
	id, err := users.NewStore(db).Register(
		users.AuthConfig{
			LegacySalt: config.LegacySalt,
			JWTSecret:  config.JWTSecret,
		},
		email,
		password)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(jsonErrorResponse{-1, err.Error()})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(map[string]int64{
		"id": id,
	})
}

// CreateContainerHandler allows creation of a container from a POST method
// Expected body:
//   name
func CreateContainerHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	user, err := users.NewStore(db).ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "User specified not found."})
		return
	}
	container := containers.Container{
		User: user,
		Name: req.PostFormValue("name"),
	}
	err = containers.NewStore(db).Create(&container)
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

// UpdateContainerHandler exposes updating a container
// @todo add support for updating the location ID
func UpdateContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	containerModel := containers.NewStore(db)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containerModel.ByID(int64(containerID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "Container not found."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(jsonErrorResponse{-2, "Not allowed to edit this container."})
		return
	}
	if name := req.PostFormValue("name"); name != "" {
		container.Name = name
	}
	err = containerModel.Update(&container)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(jsonErrorResponse{-3, "Error updating the container."})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}

func DeleteContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	containerModel := containers.NewStore(db)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containerModel.ByID(int64(containerID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "Container not found."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(jsonErrorResponse{-2, "Not allowed to edit this container."})
		return
	}
	err = containerModel.Delete(int64(containerID))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(jsonErrorResponse{-3, "Error deleting container."})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}

// ContainerHandler gets a specific container by ID
func ContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containers.NewStore(db).ByID(int64(containerID))
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
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	userModel := users.NewStore(db)
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
	limit.SetPage(page, containers.QueryLimit)
	containerModel := containers.NewStore(db)
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

// ContainerQR will output a QR code png for a specific container.
func ContainerQR(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	// @todo Figure out where this will direct to in the SPA.
	qrBytes, _ := qrcode.Encode(fmt.Sprintf("%v/container/%v", config.WebHost, vars["id"]), qrcode.Medium, 250)
	res.Write(qrBytes)
}

// SaveContainerItemHandler allows creation of a container from a POST method
// Expected body:
//   body
//   quantity
func SaveContainerItemHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	jsonOut := json.NewEncoder(res)
	vars := mux.Vars(req)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containers.NewStore(db).ByID(int64(containerID))
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(jsonErrorResponse{-1, "Failed to retrieve the container."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(jsonErrorResponse{-2, "Not allowed to modify this container."})
		return
	}
	itemModel := items.NewStore(db)
	quantity, _ := strconv.Atoi(req.PostFormValue("quantity"))
	var item items.ContainerItem
	if _, ok := vars["item_id"]; ok {
		itemID, _ := strconv.Atoi(vars["item_id"])
		item, err = itemModel.ByID(int64(itemID))
		if err != nil {
			jsonOut.Encode(jsonErrorResponse{-3, "Unable to retrieve item to modify."})
		}
	} else {
		item = items.ContainerItem{
			Container: &container,
		}
	}
	if quantity > 0 {
		item.Quantity = quantity
	}
	if body := req.PostFormValue("body"); body != "" {
		item.Body = body
	}
	if _, ok := vars["item_id"]; ok {
		itemID, _ := strconv.Atoi(vars["item_id"])
		item.ID = int64(itemID)
		err = itemModel.Update(item)
	} else {
		err = itemModel.Create(&item)
	}
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(jsonErrorResponse{-4, "Unable to create container item"})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(map[string]int64{
		"id": item.ID,
	})
}

// ContainerItemsHandler is an interface into items of a container
// @todo Consider syncing some of the non-related queries to go routines
func ContainerItemsHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	jsonOut := json.NewEncoder(res)
	vars := mux.Vars(req)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containers.NewStore(db).ByID(int64(containerID))
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "Container not found."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(jsonErrorResponse{-2, "Not allowed to view items in this container."})
		return
	}
	params := req.URL.Query()
	var limit models.QueryLimit
	page, _ := strconv.Atoi(params.Get("page"))
	limit.SetPage(page, containers.QueryLimit)
	itemModel := items.NewStore(db)
	sort := itemModel.GetSortBy(params.Get("sort_field"), models.SortType(params.Get("sort_dir")))
	response, err := itemModel.GetContainerItems(&container, sort, limit)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(jsonErrorResponse{-3, "Unable to retrieve container items."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(response)
}

// DeleteContainerItemHandler will remove an item from a container and update the container count.
func DeleteContainerItemHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	jsonOut := json.NewEncoder(res)
	vars := mux.Vars(req)
	itemID, _ := strconv.Atoi(vars["item_id"])
	itemModel := items.NewStore(db)
	item, err := itemModel.ByID(int64(itemID))
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "Item not found."})
		return
	}
	if item.Container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(jsonErrorResponse{-2, "Not allowed to delete this item."})
		return
	}
	err = itemModel.Delete(item)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(jsonErrorResponse{-3, "Unable to delete this item."})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}

// CreateLocationHandler will create a location from user input
// Expected body:
//   - name
//   - address
func CreateLocationHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	user, err := users.NewStore(db).ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "Unable to find user to associate this location."})
		return
	}
	location := locations.Location{
		User:    user,
		Name:    req.PostFormValue("name"),
		Address: req.PostFormValue("address"),
	}
	err = locations.NewStore(db).Create(&location)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(jsonErrorResponse{-2, "Unable to store location."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(map[string]int64{
		"id": location.ID,
	})
}

// UpdateLocationHandler will handle updating location based on user input
func UpdateLocationHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := GetDBResource()
	defer db.Close()
	var userKey userKey = "user"
	userID := int64(req.Context().Value(userKey).(jwt.MapClaims)["id"].(float64))
	locationModel := locations.NewStore(db)
	locationID, _ := strconv.Atoi(vars["id"])
	location, err := locationModel.ByID(int64(locationID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(jsonErrorResponse{-1, "Location not found."})
		return
	}
	if userID != location.User.ID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(jsonErrorResponse{-2, "Not allowed to modify this location."})
		return
	}
	location.Name = req.PostFormValue("name")
	location.Address = req.PostFormValue("address")
	err = locationModel.Update(&location)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(jsonErrorResponse{-3, "Failed to update location."})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}
