package authn

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeTokenVerifier struct {
	identity TokenIdentity
	err      error
	token    string
}

func (v *fakeTokenVerifier) VerifyIDToken(
	_ context.Context,
	token string,
) (TokenIdentity, error) {
	v.token = token
	return v.identity, v.err
}

type fakeUserResolver struct {
	userID      int64
	err         error
	uid         string
	displayName string
}

func (r *fakeUserResolver) ResolveFirebaseUser(
	_ context.Context,
	uid string,
	displayName string,
) (int64, error) {
	r.uid = uid
	r.displayName = displayName
	return r.userID, r.err
}

func TestFirebaseMiddlewareAuthenticatesUser(t *testing.T) {
	verifier := &fakeTokenVerifier{identity: TokenIdentity{
		UID:         "firebase-user-1",
		DisplayName: "テストユーザー",
	}}
	resolver := &fakeUserResolver{userID: 42}
	middleware := NewFirebaseMiddleware(verifier, resolver)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/settings", nil)
	request.Header.Set("Authorization", "Bearer id-token")
	response := httptest.NewRecorder()

	middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatal("authenticated user missing from context")
		}
		if user.ID != 42 || user.FirebaseUID != "firebase-user-1" {
			t.Fatalf("user = %#v, want ID 42 and Firebase UID", user)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if verifier.token != "id-token" {
		t.Fatalf("verified token = %q, want id-token", verifier.token)
	}
	if resolver.uid != "firebase-user-1" || resolver.displayName != "テストユーザー" {
		t.Fatalf("resolver received UID %q and name %q", resolver.uid, resolver.displayName)
	}
}

func TestFirebaseMiddlewareRejectsMissingOrInvalidToken(t *testing.T) {
	tests := []struct {
		name       string
		authorizer string
		verifyErr  error
	}{
		{name: "missing header"},
		{name: "wrong scheme", authorizer: "Basic token"},
		{name: "invalid token", authorizer: "Bearer invalid", verifyErr: errors.New("invalid")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			verifier := &fakeTokenVerifier{
				identity: TokenIdentity{UID: "firebase-user-1"},
				err:      test.verifyErr,
			}
			middleware := NewFirebaseMiddleware(verifier, &fakeUserResolver{userID: 1})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/me/settings", nil)
			if test.authorizer != "" {
				request.Header.Set("Authorization", test.authorizer)
			}
			response := httptest.NewRecorder()

			middleware.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("next handler must not be called")
			})).ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			if response.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Error("WWW-Authenticate header is missing")
			}
		})
	}
}

func TestFirebaseMiddlewareHandlesUserResolutionFailure(t *testing.T) {
	middleware := NewFirebaseMiddleware(
		&fakeTokenVerifier{identity: TokenIdentity{UID: "firebase-user-1"}},
		&fakeUserResolver{err: errors.New("database unavailable")},
	)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/settings", nil)
	request.Header.Set("Authorization", "Bearer id-token")
	response := httptest.NewRecorder()

	middleware.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	})).ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}

func TestDevMiddlewareUsesConfiguredUserWithoutToken(t *testing.T) {
	middleware := NewDevMiddleware(7)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/settings", nil)
	response := httptest.NewRecorder()

	middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok || user.ID != 7 {
			t.Fatalf("user = %#v, ok = %t; want user 7", user, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
