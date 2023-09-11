package server

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Session struct {
	ID    string
	Guest bool
	jwt.RegisteredClaims
}

func NewSession(id string) Session {
	return Session{
		ID:               id,
		Guest:            false,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
}

func GuestSession() Session {
	session := NewSession(uuid.New().String())
	session.Guest = true
	return session
}

type AuthServer struct {
	jwtKey string
}

func NewAuthServer(jwtKey string) *AuthServer {
	return &AuthServer{jwtKey}
}

func (server *AuthServer) keyFunc(token *jwt.Token) (interface{}, error) {
	return server.jwtKey, nil
}

func (server *AuthServer) SetSession(w http.ResponseWriter, session Session) error {
	expiry := time.Now().Add(24 * time.Hour)

	session.ExpiresAt = jwt.NewNumericDate(expiry)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, session)
	tokenString, err := token.SignedString(server.jwtKey)
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

func (server *AuthServer) GetSession(w http.ResponseWriter, r *http.Request) (*Session, error) {
	cookie, err := r.Cookie("session")

	var session Session
	if cookie != nil && err == nil {
		token, err := jwt.ParseWithClaims(cookie.Value, &session, server.keyFunc)
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

	server.SetSession(w, session)
	return &session, nil
}

func (server *AuthServer) Login(w http.ResponseWriter, r *http.Request) {

}

func (server *AuthServer) Logout(w http.ResponseWriter, r *http.Request) {

}
