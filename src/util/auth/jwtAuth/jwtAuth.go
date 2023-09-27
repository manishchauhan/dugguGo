package jwtAuth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/manishchauhan/dugguGo/servers/errorhandler"
)

const (
	refreshTokenCookieName = "refresh_token"
	accessTokenCookieName  = "access_token"
	refreshTokenDuration   = 24 * 60 * time.Hour // Refresh token expiration: 60 days
	accessTokenDuration    = 15 * time.Minute    // Access token expiration: 15 minutes
)

var (
	userIDContextKey   = "userid"
	usernameContextKey = "username"
	emailContextKey    = "email"
)

var jwtSecret []byte

// CustomClaims represents custom claims to be included in the JWT token.
type CustomClaims struct {
	UserID   int    `json:"userid"`
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

// validate access token secret and return claims or error
func ParseAndValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, _ := token.Claims.(*CustomClaims)
	return claims, nil
}

// Each route should contain this middle which would confrim that user need to be authenticated or
// not based on jwt toekn
// AuthMiddleware is a middleware to authenticate and authorize requests using JWT.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to retrieve the access and refresh tokens from cookies
		accessToken, accessTokenErr := r.Cookie(accessTokenCookieName)
		refreshToken, refreshTokenErr := r.Cookie(refreshTokenCookieName)

		// Check if both access token and refresh token are missing
		if accessTokenErr == http.ErrNoCookie && refreshTokenErr == http.ErrNoCookie {
			errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Session is invalid or has expired. Please log in again.")
			return
		}

		if accessTokenErr == nil {
			// Access token is present, attempt to validate it
			claims, err := ParseAndValidateToken(accessToken.Value)
			if err != nil {
				if refreshTokenErr == nil {
					// Access token is invalid, try refreshing with the refresh token
					claims, err := ParseAndValidateToken(refreshToken.Value)
					if err != nil {
						errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Session is invalid or has expired. Please log in again.")
						return
					}

					// Generate a new access token and set it in the response cookie
					newAccessToken, err := CreateAccessToken(claims.UserID, claims.UserName, claims.Email)
					if err != nil {
						http.Error(w, "Error creating access token", http.StatusInternalServerError)
						errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Error creating access token")
						return
					}
					SetAccessTokenInCookie(w, newAccessToken)

					// Store user information in the request context
					ctx := addToContext(r.Context(),
						userIDContextKey, claims.UserID,
						usernameContextKey, claims.UserName,
						emailContextKey, claims.Email,
					)

					// Proceed with the next HTTP handler
					next.ServeHTTP(w, r.WithContext(ctx))
				} else {
					errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Access token is invalid. Please log in again.")
				}
				return
			}

			// Access token is valid, store user information in the request context
			fmt.Println(claims)
			ctx := addToContext(r.Context(),
				userIDContextKey, claims.UserID,
				usernameContextKey, claims.UserName,
				emailContextKey, claims.Email,
			)

			// Proceed with the next HTTP handler
			next.ServeHTTP(w, r.WithContext(ctx))
		} else if refreshTokenErr == nil {
			// Access token is missing but refresh token is present, refresh it
			claims, err := ParseAndValidateToken(refreshToken.Value)
			if err != nil {
				errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "Session is invalid or has expired. Please log in again.")
				return
			}

			// Generate a new access token and set it in the response cookie
			newAccessToken, err := CreateAccessToken(claims.UserID, claims.UserName, claims.Email)
			if err != nil {
				http.Error(w, "Error creating access token", http.StatusInternalServerError)
				errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Error creating access token")
				return
			}
			SetAccessTokenInCookie(w, newAccessToken)

			// Store user information in the request context
			ctx := addToContext(r.Context(),
				userIDContextKey, claims.UserID,
				usernameContextKey, claims.UserName,
				emailContextKey, claims.Email,
			)

			// Proceed with the next HTTP handler
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

// Helper function to add multiple key-value pairs to the context
func addToContext(ctx context.Context, keyValues ...interface{}) context.Context {
	for i := 0; i < len(keyValues); i += 2 {
		key := keyValues[i]
		value := keyValues[i+1]
		ctx = context.WithValue(ctx, key, value)
	}

	return ctx
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

func SetCookies(w http.ResponseWriter, newAccessToken string, newRefreshToken string) {
	SetAccessTokenInCookie(w, newAccessToken)
	SetRefreshTokenCookie(w, newRefreshToken)
}
func SetAccessTokenInCookie(w http.ResponseWriter, newAccessToken string) {
	http.SetCookie(w, &http.Cookie{Name: accessTokenCookieName, Secure: true, Value: newAccessToken, Path: "/", Domain: "localhost", HttpOnly: false, MaxAge: int(accessTokenDuration.Seconds()), SameSite: http.SameSiteNoneMode})
}
func SetRefreshTokenCookie(w http.ResponseWriter, newRefreshToken string) {
	http.SetCookie(w, &http.Cookie{Name: refreshTokenCookieName, Secure: true, Value: newRefreshToken, Path: "/", Domain: "localhost", HttpOnly: false, MaxAge: int(refreshTokenDuration.Seconds()), SameSite: http.SameSiteNoneMode})
}

// Helper function to clear cookies
func ClearCookies(w http.ResponseWriter) {
	// Clear the access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     accessTokenCookieName,
		Secure:   true,
		Value:    "",
		Path:     "/",
		Domain:   "localhost", // Replace with your domain
		HttpOnly: false,
		MaxAge:   0, // Set MaxAge to 0 to delete the cookie
		SameSite: http.SameSiteNoneMode,
	})

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Secure:   true,
		Value:    "",
		Path:     "/",
		Domain:   "localhost", // Replace with your domain
		HttpOnly: false,
		MaxAge:   0, // Set MaxAge to 0 to delete the cookie
		SameSite: http.SameSiteNoneMode,
	})
}
