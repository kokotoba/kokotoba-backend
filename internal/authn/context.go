package authn

import "context"

type User struct {
	ID          int64
	FirebaseUID string
}

type userContextKey struct{}

func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey{}).(User)
	return user, ok && user.ID > 0
}
