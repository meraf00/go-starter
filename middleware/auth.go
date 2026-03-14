package middleware

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/chariotplatform/goapi/transport"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var ErrUnauthorized = transport.Unauthorized("Unauthorized", "AUTH_01", nil)

type authContextKey int

const (
	authContext authContextKey = iota
)

type AuthContext struct {
	UID         uuid.UUID
	Permissions []string
}

func WithAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContext, authCtx)
}

func GetAuthContext(ctx context.Context) (*AuthContext, error) {
	authCtx, ok := ctx.Value(authContext).(*AuthContext)
	if !ok || authCtx == nil {
		return nil, errors.New("unauthorized: missing auth context")
	}

	return authCtx, nil
}

// Auth middleware for Ory Kratos integration
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// cookieHeader := r.Header.Get("Cookie")

		// Set auth context
		ctx := WithAuthContext(r.Context(), &AuthContext{})

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission is a middleware that checks if the authenticated user has the required permissions.
// It must be used after AuthenticationMiddleware to ensure the AuthContext is present.
func RequirePermission(requiredPermissions ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := GetAuthContext(r.Context())
			if err != nil {
				transport.Error(w, ErrUnauthorized)
				return
			}

			for _, requiredPermission := range requiredPermissions {
				if !slices.Contains(ctx.Permissions, string(requiredPermission)) {
					transport.Error(w, ErrUnauthorized)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
