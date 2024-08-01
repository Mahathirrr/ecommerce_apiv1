package handler

import (
	"context"
	"ecom_apiv1/token"
	"fmt"
	"net/http"
	"strings"
)

type authKey struct{}

func GetAuthMiddlewareFunc(tokenMaker *token.JWTMaker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read authorization header
			// verify token
			claims, err := verifyClaimsFromHeader(r, tokenMaker)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error verifying token: %v", err), http.StatusUnauthorized)
				return
			}
			// pass the payload/claims down the context
			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAdminMiddlewareFunc(tokenMaker *token.JWTMaker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read authorization header
			// verify token
			claims, err := verifyClaimsFromHeader(r, tokenMaker)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error verifying token: %v", err), http.StatusUnauthorized)
				return
			}
			if !claims.IsAdmin {
				http.Error(w, "user is not admin", http.StatusForbidden)
				return
			}
			// pass the payload/claims down the context
			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func verifyClaimsFromHeader(r *http.Request, tokenMaker *token.JWTMaker) (*token.UserClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authrorization header is missing")
	}
	field := strings.Fields(authHeader)
	if len(field) != 2 || field[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authroziation header")
	}
	token := field[1]
	claims, err := tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}
