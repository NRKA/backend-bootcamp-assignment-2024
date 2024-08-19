package auth

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenNotValidYet = errors.New("token not valid yet")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrInvalidPassword  = errors.New("invalid password")
)
