package authn

import (
	"context"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
)

type FirebaseTokenVerifier struct {
	client *firebaseauth.Client
}

func NewFirebaseTokenVerifier(
	ctx context.Context,
	projectID string,
) (*FirebaseTokenVerifier, error) {
	var config *firebase.Config
	if projectID != "" {
		config = &firebase.Config{ProjectID: projectID}
	}
	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return &FirebaseTokenVerifier{client: client}, nil
}

func (v *FirebaseTokenVerifier) VerifyIDToken(
	ctx context.Context,
	rawToken string,
) (TokenIdentity, error) {
	token, err := v.client.VerifyIDToken(ctx, rawToken)
	if err != nil {
		return TokenIdentity{}, err
	}
	displayName, _ := token.Claims["name"].(string)
	return TokenIdentity{UID: token.UID, DisplayName: displayName}, nil
}
