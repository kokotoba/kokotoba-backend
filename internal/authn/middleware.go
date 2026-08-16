package authn

import (
	"context"
	"log"
	"net/http"
	"strings"

	"kokotoba-backend/internal/model"
	"kokotoba-backend/internal/view"
)

type TokenIdentity struct {
	UID         string
	DisplayName string
}

type TokenVerifier interface {
	VerifyIDToken(context.Context, string) (TokenIdentity, error)
}

type UserResolver interface {
	ResolveFirebaseUser(context.Context, string, string) (int64, error)
}

type Middleware struct {
	mode      Mode
	devUserID int64
	verifier  TokenVerifier
	resolver  UserResolver
}

func NewFirebaseMiddleware(verifier TokenVerifier, resolver UserResolver) *Middleware {
	return &Middleware{
		mode:     ModeFirebase,
		verifier: verifier,
		resolver: resolver,
	}
}

func NewDevMiddleware(userID int64) *Middleware {
	return &Middleware{mode: ModeDev, devUserID: userID}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.mode == ModeDev {
			next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), User{
				ID: m.devUserID,
			})))
			return
		}

		rawToken, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			unauthorized(w)
			return
		}
		identity, err := m.verifier.VerifyIDToken(r.Context(), rawToken)
		if err != nil || strings.TrimSpace(identity.UID) == "" {
			unauthorized(w)
			return
		}
		userID, err := m.resolver.ResolveFirebaseUser(
			r.Context(),
			identity.UID,
			identity.DisplayName,
		)
		if err != nil {
			log.Printf("resolve authenticated user: %v", err)
			_ = view.JSON(w, http.StatusInternalServerError, model.ErrorResponse{
				Message: "failed to resolve authenticated user",
			})
			return
		}

		next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), User{
			ID:          userID,
			FirebaseUID: identity.UID,
		})))
	})
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	_ = view.JSON(w, http.StatusUnauthorized, model.ErrorResponse{
		Message: "valid Firebase ID token is required",
	})
}
