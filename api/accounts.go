package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jorwong/go_user_accounts/models"
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

	if !ratelimit.IsAllowed(foundUser.Email) {
		pkg.LogChannel <- time.Now().String() + "," + "Rate Limited for " + credentials.Email

		http.Error(w, "Rate Limited", http.StatusTooManyRequests)
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
	pkg.LogChannel <- time.Now().String() + "," + "Successful Login for " + foundUser.Email
}

func createToken(email string) string {
	h := sha256.New()
	h.Write([]byte(time.Now().String() + email))
	token := hex.EncodeToString(h.Sum(nil))

	return token
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

	// check if upon calling is the session valid?

	message, result := CheckIfUserSessionIsValid(credentials.Session, credentials.Email)

	if result == false {
		switch message {
		case ERROR:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		case INVALID_SESSION:
			http.Error(w, "Invalid Session", http.StatusBadRequest)
			return
		}
	}

	User, err := models.FindUserByEmail(credentials.Email)
	if err != nil || User == nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(User.ToString()))
	w.WriteHeader(http.StatusOK)
}
