package main

import (
	"github.com/gorilla/mux"
	"github.com/jorwong/go_user_accounts/api"
	"github.com/jorwong/go_user_accounts/models"
	jwt "github.com/jorwong/go_user_accounts/pkg/jwt"
	pkg "github.com/jorwong/go_user_accounts/pkg/logging"
	"net/http"
)

func main() {
	models.InitDB()
	pkg.StartLoggerWorker()
	router := mux.NewRouter()

	router.HandleFunc("/register", api.Register).Methods("POST")
	router.HandleFunc("/login", api.Login).Methods("POST")
	router.HandleFunc("/logout", api.Logout).Methods("POST")

	authenticationSubrouter := router.PathPrefix("/auth").Subrouter()
	authenticationSubrouter.Use(jwt.VerifyJWT)
	authenticationSubrouter.HandleFunc("/profile", api.GetProfile).Methods("POST")

	http.ListenAndServe(":8080", router)
}
