package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const jwtTokenExpiry = time.Minute * 15
const refreshTokenExpiry = time.Hour * 24

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserName string `json:"name"`
	jwt.RegisteredClaims
}

func (app *application) getTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	//	add a header
	w.Header().Add("Valy", "Authorization")

	//	get the authorization header
	authHeader := r.Header.Get("Authorization")

	//	sanity check
	if authHeader == "" {
		return "", nil, errors.New("no auth header")
	}

	//	split a header on spaces and extract token
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", nil, errors.New("invalid auth header")
	}
	token := headerParts[1]

	//	declare and empty Claims var and fill it in by content of token
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		//	validate the signing method algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(app.JWTSecret), nil
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}
		return "", nil, err
	}

	//	check it token is issued by us
	if claims.Issuer != app.Domain {
		return "", nil, errors.New("incorrect issuer")
	}

	//	valid token
	return token, claims, nil
}
