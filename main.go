package main

import (
	"github.com/gorilla/mux"
	"github.com/jorwong/go_user_accounts/api"
	"github.com/jorwong/go_user_accounts/models"
	"net/http"
)

func main() {
	models.InitDB()

	router := mux.NewRouter()

	router.HandleFunc("/register", api.Register).Methods("POST")
	router.HandleFunc("/login", api.Login).Methods("POST")
	router.HandleFunc("/logout", api.Logout).Methods("POST")
	router.HandleFunc("/profile", api.GetProfile).Methods("POST")
	http.ListenAndServe(":8080", router)
}
