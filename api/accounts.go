package api

import (
	"encoding/json"
	"fmt"
	"github.com/jorwong/go_user_accounts/models"
	jwt "github.com/jorwong/go_user_accounts/pkg/jwt"
	pkg "github.com/jorwong/go_user_accounts/pkg/logging"
	ratelimit "github.com/jorwong/go_user_accounts/pkg/ratelimit"
	"io"
	"net/http"
	"time"
)

func Register(w http.ResponseWriter, req *http.Request) {
	form := req.Body

	var registerForm struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(form)
	err := decoder.Decode(&registerForm)
	defer req.Body.Close() // <-- FIX: Ensure the request body stream is closed

	if registerForm.Name == "" || registerForm.Email == "" || registerForm.Password == "" {
		// If any required field is empty
		http.Error(w, "Missing required fields (Name, Email, or Password).", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = models.CreateUser(registerForm.Email, registerForm.Name, registerForm.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	responseMessage := "User registered successfully!"
	_, err = w.Write([]byte(responseMessage))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, req *http.Request) {
	form := req.Body

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(form)
	err := decoder.Decode(&credentials)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(req.Body)

	if err != nil {
		pkg.LogChannel <- time.Now().String() + "," + "Bad Request: " + err.Error()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if credentials.Email == "" || credentials.Password == "" {
		pkg.LogChannel <- time.Now().String() + "," + "Missing required fields (Email, or Password)."
		http.Error(w, "Missing required fields (Email, or Password).", http.StatusBadRequest)
		return
	}

	foundUser, err := models.FindUserByEmail(credentials.Email)

	if err != nil && err.Error() == "DB_ERROR" {
		pkg.LogChannel <- time.Now().String() + "," + "DB ERROR"
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	if foundUser == nil || !foundUser.CheckPasswordHash(credentials.Password) {
		pkg.LogChannel <- time.Now().String() + "," + "Invalid Credentials for " + credentials.Email

		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	ifIsAllowed, err := ratelimit.IsAllowed(foundUser.Email)
	if err != nil && err.Error() != "RATE_LIMITED" {
		pkg.LogChannel <- time.Now().String() + "," + err.Error()
		http.Error(w, "ERROR", http.StatusInternalServerError)
		return
	}
	if !ifIsAllowed {
		pkg.LogChannel <- time.Now().String() + "," + "Rate Limited for " + credentials.Email

		http.Error(w, "Rate Limited", http.StatusTooManyRequests)
		return
	}

	//timeExpire := time.Now().Add(time.Duration(time.Second) * 60) // expire after 60 seconds

	jwtToken, err := jwt.GenerateJWT(foundUser.Email)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("TOKEN: " + jwtToken))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg.LogChannel <- time.Now().String() + "," + "Successful Login for " + foundUser.Email
}

func Logout(w http.ResponseWriter, req *http.Request) {
	form := req.Body

	var credentials struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(form)
	err := decoder.Decode(&credentials)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(req.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if credentials.Email == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	foundUser, err := models.FindUserByEmail(credentials.Email)

	if err != nil && err.Error() == "DB_ERROR" {
		http.Error(w, "DB Error", http.StatusInternalServerError)
	}

	err = models.RevokeSession(foundUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetProfile(w http.ResponseWriter, req *http.Request) {
	form := req.Body

	var credentials struct {
		Session string `json:"session"`
		Email   string `json:"email"`
	}

	decoder := json.NewDecoder(form)
	err := decoder.Decode(&credentials)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(req.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if credentials.Session == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	User, err := models.FindUserByEmail(credentials.Email)
	if err != nil || User == nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(User.ToString()))
	w.WriteHeader(http.StatusOK)
}
