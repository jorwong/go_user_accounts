package main

import (
	"github.com/gorilla/mux"
	"github.com/jorwong/go_user_accounts/models"
	"net/http"
)

func main() {
	models.InitDB()

	router := mux.NewRouter()

	router.HandleFunc("/hello", hello).Methods("GET")
	router.HandleFunc("/headers", headers).Methods("GET")
	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")

	http.ListenAndServe(":8080", router)
}
