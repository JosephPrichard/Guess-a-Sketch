/*
 * Copyright (c) Joseph Prichard 2023
 */

package servers

import (
	"errors"
	"fmt"
	"guessthesketch/game"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Authenticator interface {
	GetSession(token string) (*JwtSession, error)
	GetPlayer(token string) game.Player
}

type AuthServer struct {
	jwtKey []byte
}

func NewAuthServer(jwtKey string) *AuthServer {
	return &AuthServer{jwtKey: []byte(jwtKey)}
}

func (server *AuthServer) keyFunc(_ *jwt.Token) (interface{}, error) {
	return server.jwtKey, nil
}

func (server *AuthServer) GenerateToken(session JwtSession) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, session)
	tokenString, err := token.SignedString(server.jwtKey)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to generate token for session %s with error %s", session.ID, err.Error()))
	}
	return tokenString, nil
}

// gets the session from a request, returning an error if it cannot be extracted or a nil session if there is no session
func (server *AuthServer) GetSession(token string) (*JwtSession, error) {
	var session JwtSession
	if token != "" {
		jwtToken, err := jwt.ParseWithClaims(token, &session, server.keyFunc)
		if err != nil {
			log.Printf("Failed to parse jwt with error %s", err.Error())
			return nil, err
		}
		if !jwtToken.Valid {
			return nil, nil
		}
	} else {
		return nil, nil
	}

	return &session, nil
}

func (server *AuthServer) GetPlayer(token string) game.Player {
	player := GuestUser()
	if token != "" {
		// if a session token is specified, attempt to get the id for the user
		session, err := server.GetSession(token)
		if err != nil && session != nil {
			player = session.user
		}
	}
	return player
}

func (server *AuthServer) EstablishSession(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	token := r.Header.Get("token")
	session, err := server.GetSession(token)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if session == nil {
		newSession := NewSession(GuestUser(), true)
		session = &newSession
		token, err = server.GenerateToken(newSession)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	log.Printf("JwtSession with id %s", session.ID)

	type TokenResp struct {
		Token string `json:"token"`
	}
	tokenResp := TokenResp{Token: token}
	w.WriteHeader(http.StatusOK)
	WriteJson(w, tokenResp)
	return
}

func (server *AuthServer) Login(w http.ResponseWriter, r *http.Request) {

}

func (server *AuthServer) Logout(w http.ResponseWriter, r *http.Request) {

}

type JwtSession struct {
	user  game.Player
	Guest bool
	jwt.RegisteredClaims
}

func NewSession(user game.Player, isGuest bool) JwtSession {
	expiry := time.Now().Add(24 * time.Hour)
	claims := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expiry)}
	return JwtSession{
		user:             user,
		Guest:            isGuest,
		RegisteredClaims: claims,
	}
}

func GuestUser() game.Player {
	return game.Player{ID: uuid.New(), Name: fmt.Sprintf("Guest %d", 10+rand.Intn(89))}
}
