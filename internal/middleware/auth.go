package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Auth is the middleware factory. It needs the secret key to verify the signature.
func Auth(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			
			// 1. Look for the "Authorization" header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
				return
			}

			// 2. The standard format is "Bearer <token>". We need to split it.
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			// 3. Parse and validate the token using our secret key
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Verify the algorithm is what we expect
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// 4. Extract the data (Claims) from the passport
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// JSON numbers are parsed as float64, so we cast to float64, then to int
			userID := int(claims["user_id"].(float64))

			// 5. THE MAGIC TRICK: Put the user_id into the Request "Context"
			// This allows us to pass the user_id down to the actual handler!
			ctx := context.WithValue(r.Context(), "user_id", userID)
			reqWithContext := r.WithContext(ctx)

			// 6. Let them pass! (Handing over the modified request)
			next.ServeHTTP(w, reqWithContext)
		})
	}
}