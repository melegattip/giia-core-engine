package domain

import "errors"

// Common domain errors
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserInactive          = errors.New("user is inactive")
	ErrRoleNotFound          = errors.New("role not found")
	ErrRoleAlreadyExists     = errors.New("role already exists")
	ErrPermissionNotFound    = errors.New("permission not found")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrOrganizationNotFound  = errors.New("organization not found")
	ErrInvalidToken          = errors.New("invalid token")
	ErrTokenExpired          = errors.New("token expired")
	ErrCircularRoleHierarchy = errors.New("circular role hierarchy detected")
)
