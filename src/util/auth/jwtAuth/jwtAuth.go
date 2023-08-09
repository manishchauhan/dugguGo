package jwtAuth

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	refreshTokenCookieName = "refresh_token"
	accessTokenCookieName  = "access_token"
	refreshTokenDuration   = 24 * 30 * time.Hour // Refresh token expiration: 30 days
	accessTokenDuration    = 15 * time.Minute    // Access token expiration: 15 minutes
)

type contextKey string

const userContextKey contextKey = "user"

var jwtSecret []byte

// CustomClaims represents custom claims to be included in the JWT token.
type CustomClaims struct {
	UserID   string `json:"id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	jwt.StandardClaims
}

// This function should be called after the .env file has been loaded.
func SetEnvData() {
	jwtSecret = []byte(os.Getenv(
		"ACCESS_TOKEN_SECRET",
	))
}

// Each route should contain this middle which would confrim that user need to be authenticated or
// not based on jwt toekn
// AuthMiddleware is a middleware to authenticate and authorize requests using JWT.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie(accessTokenCookieName)
		if err != nil {
			http.Error(w, "Access token missing", http.StatusUnauthorized)
			return
		}

		tokenString := accessToken.Value

		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			refreshToken, err := r.Cookie(refreshTokenCookieName)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			refreshTokenString := refreshToken.Value
			claims := &CustomClaims{}
			refreshTokenToken, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
			if err != nil || !refreshTokenToken.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Refresh access token and set new access token and refresh token cookies
			newAccessToken := createAccessToken(claims.UserID, claims.UserName, claims.Email)
			newRefreshToken := createRefreshToken(claims.UserID, claims.UserName, claims.Email)
			http.SetCookie(w, &http.Cookie{Name: accessTokenCookieName, Value: newAccessToken, HttpOnly: true, MaxAge: int(accessTokenDuration.Seconds())})
			http.SetCookie(w, &http.Cookie{Name: refreshTokenCookieName, Value: newRefreshToken, HttpOnly: true, MaxAge: int(refreshTokenDuration.Seconds())})

			r = r.WithContext(context.WithValue(r.Context(), userContextKey, claims.UserID))
			next.ServeHTTP(w, r)
		} else {
			claims, _ := token.Claims.(*CustomClaims)
			r = r.WithContext(context.WithValue(r.Context(), userContextKey, claims.UserID))
			next.ServeHTTP(w, r)
		}
	})
}

func createAccessToken(userid, username, email string) string {
	claims := &CustomClaims{
		UserID:   userid,
		UserName: username,
		Email:    email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(accessTokenDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}

func createRefreshToken(userid, username, email string) string {
	claims := &CustomClaims{
		UserID:   userid,
		UserName: username,
		Email:    email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(refreshTokenDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}
