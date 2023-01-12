package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"webapp/pkg/data"

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
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
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

func (app *application) generateTokenPair(user *data.User) (TokenPairs, error) {
	//	create the token
	token := jwt.New(jwt.SigningMethodHS256)

	//	set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = app.Domain
	claims["iss"] = app.Domain
	if user.IsAdmin == 1 {
		claims["admin"] = true
	} else {
		claims["admin"] = false
	}
	//	set the expiry
	claims["exp"] = time.Now().Add(jwtTokenExpiry).Unix()
	//	create the signed token
	signedAccessToken, err := token.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}
	//	create the refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["exp"] = time.Now().Add(refreshTokenExpiry).Unix()
	//	create signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(app.JWTSecret))
	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}
	return tokenPairs, nil
}
