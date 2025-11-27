package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jorwong/go_user_accounts/models"
	"io"
	"net/http"
	"time"
)

func hello(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	fmt.Println("server: hello handler started")
	defer fmt.Println("server: hello handler ended")

	select {
	case <-ctx.Done():
		err := ctx.Err()
		fmt.Println("server:", err)
		internalError := http.StatusInternalServerError
		http.Error(w, err.Error(), internalError)
	case <-time.After(10 * time.Second):
		fmt.Fprintf(w, "Hello World")
	}
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func register(w http.ResponseWriter, req *http.Request) {
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

func login(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if credentials.Email == "" || credentials.Password == "" {
		http.Error(w, "Missing required fields (Email, or Password).", http.StatusBadRequest)
		return
	}

	foundUser, err := models.FindUserByEmail(credentials.Email)

	if err != nil && err.Error() == "DB_ERROR" {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	if foundUser == nil || !foundUser.CheckPasswordHash(credentials.Password) {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	timeExpire := time.Now().Add(time.Duration(time.Second) * 60) // expire after 60 seconds

	token := createToken(foundUser.Email)
	sessionToken, createdNew, err := models.CreateRedisSession(token, timeExpire, foundUser)

	if err != nil {
		fmt.Println(err)
		return
	}

	if !createdNew {
		_, err := w.Write([]byte(sessionToken))
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// Create the session for this user
	createdSession, err := models.CreateSession(foundUser, timeExpire, token)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write to redis

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(createdSession.Token))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createToken(email string) string {
	h := sha256.New()
	h.Write([]byte(time.Now().String() + email))
	token := hex.EncodeToString(h.Sum(nil))

	return token
}
