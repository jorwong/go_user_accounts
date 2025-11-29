package api

import (
	"github.com/jorwong/go_user_accounts/models"
)

const (
	INVALID_SESSION = "INVALID_SESSION"
	ERROR           = "ERROR"
	VALID           = "VALID"
)

func CheckIfUserSessionIsValid(session string, email string) (string, bool) {
	sessionKey, err := models.GetTokenFromRedis(email)
	if err != nil {
		return ERROR, false
	}

	if sessionKey != session {
		return INVALID_SESSION, false
	}

	return VALID, true
}
