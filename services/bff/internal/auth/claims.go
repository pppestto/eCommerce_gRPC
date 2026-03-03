package auth

import (
	"context"
)

type contextKey string

var ContextKeyUserID contextKey = "user_id"

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ContextKeyUserID)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
