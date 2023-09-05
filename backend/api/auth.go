package api

import (
	"guessasketch/utils"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtKey = utils.GetEnvVar("JWT_SECRET_KEY")
var keyFunc = func(token *jwt.Token) (interface{}, error) {
	return jwtKey, nil
}

type Session struct {
	ID string
	jwt.RegisteredClaims
}

func NewSession(id string) Session {
	return Session{
		ID:               id,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
}

func GuestSession() Session {
	return NewSession(uuid.New().String())
}

func SetSession(w http.ResponseWriter, session Session) error {
	expiry := time.Now().Add(24 * time.Hour)

	session.ExpiresAt = jwt.NewNumericDate(expiry)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, session)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:    "session",
		Value:   tokenString,
		Expires: expiry,
	}
	http.SetCookie(w, cookie)
	return nil
}

func GetSession(w http.ResponseWriter, r *http.Request) (*Session, error) {
	cookie, err := r.Cookie("session")

	var session Session
	if cookie != nil && err == nil {
		token, err := jwt.ParseWithClaims(cookie.Value, &session, keyFunc)
		if err != nil {
			return nil, err
		}
		if !token.Valid {
			session = GuestSession()
		}
	} else {
		// issue a new guest session
		session = GuestSession()
	}

	SetSession(w, session)
	return &session, nil
}

type AuthServer struct {
}

func (server *AuthServer) Login(w http.ResponseWriter, r *http.Request) {

}
