package api

import (
	"encoding/json"
	"guessasketch/utils"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Session struct {
	ID string
	jwt.RegisteredClaims
}

func NewSession(id string) Session {
	return Session{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
}

func SetSession(w http.ResponseWriter, session Session) error {
	expiry := time.Now().Add(24 * time.Hour)

	session.ExpiresAt = jwt.NewNumericDate(expiry)

	value, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// update the session cookie
	cookie := &http.Cookie{
		Name:    "session",
		Value:   string(value),
		Expires: expiry,
	}
	http.SetCookie(w, cookie)
	return nil
}

func GetSession(w http.ResponseWriter, r *http.Request) (*Session, error) {
	cookie, err := r.Cookie("session")

	var session Session
	if cookie != nil && err == nil {
		// unmarshal the old session
		err = json.Unmarshal([]byte(cookie.Value), &session)
		if err != nil {
			return nil, err
		}
	} else {
		// issue a new session
		session = NewSession(uuid.New().String())
	}
	
	SetSession(w, session)
	return &session, nil
}

type AuthServer struct {
}

func (server *AuthServer) Login(w http.ResponseWriter, r *http.Request) {

}
