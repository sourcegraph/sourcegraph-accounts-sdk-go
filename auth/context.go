package auth

import (
	"context"
)

type contextKey int

const (
	errorKey contextKey = iota
	userInfoKey
)

// ErrorFromContext returns the error from the given context.
func ErrorFromContext(ctx context.Context) *Error {
	return ctx.Value(errorKey).(*Error)
}

// WithError returns a new context with the given error.
func WithError(ctx context.Context, err *Error) context.Context {
	return context.WithValue(ctx, errorKey, err)
}

// UserInfoFromContext returns the user info from the given context.
func UserInfoFromContext(ctx context.Context) *UserInfo {
	return ctx.Value(userInfoKey).(*UserInfo)
}

// WithUserInfo returns a new context with the given user info.
func WithUserInfo(ctx context.Context, userInfo *UserInfo) context.Context {
	return context.WithValue(ctx, userInfoKey, userInfo)
}
