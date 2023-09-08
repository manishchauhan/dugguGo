package jwtAuth

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/manishchauhan/dugguGo/servers/errorhandler"
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
	UserID   int    `json:"id"`
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
			errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Access token missing")
			return
		}

		tokenString := accessToken.Value

		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			refreshToken, err := r.Cookie(refreshTokenCookieName)
			if err != nil {
				errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			refreshTokenString := refreshToken.Value
			claims := &CustomClaims{}
			refreshTokenToken, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
			if err != nil || !refreshTokenToken.Valid {
				errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Refresh access token and set new access token and refresh token cookies
			newAccessToken, accessTokenError := CreateAccessToken(claims.UserID, claims.UserName, claims.Email)
			if err != accessTokenError {
				errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			newRefreshToken, refreshTokenError := CreateRefreshToken(claims.UserID, claims.UserName, claims.Email)
			if err != refreshTokenError {
				errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			SetCookie(w, newAccessToken, newRefreshToken)
			r = r.WithContext(context.WithValue(r.Context(), userContextKey, claims.UserID))
			next.ServeHTTP(w, r)
		} else {
			claims, _ := token.Claims.(*CustomClaims)
			r = r.WithContext(context.WithValue(r.Context(), userContextKey, claims.UserID))
			next.ServeHTTP(w, r)
		}
	})
}

func CreateAccessToken(userid int, username string, email string) (string, error) {
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
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func CreateRefreshToken(userid int, username string, email string) (string, error) {
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
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func SetCookie(w http.ResponseWriter, newAccessToken string, newRefreshToken string) {
	http.SetCookie(w, &http.Cookie{Name: accessTokenCookieName, Secure: true, Value: newAccessToken, Path: "/", Domain: "localhost", HttpOnly: false, MaxAge: int(accessTokenDuration.Seconds()), SameSite: http.SameSiteNoneMode})
	http.SetCookie(w, &http.Cookie{Name: refreshTokenCookieName, Secure: true, Value: newRefreshToken, Path: "/", Domain: "localhost", HttpOnly: false, MaxAge: int(refreshTokenDuration.Seconds()), SameSite: http.SameSiteNoneMode})
}
