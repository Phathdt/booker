package interceptors

import "context"

type contextKey string

const (
	userIDKey contextKey = "user_id"
	roleKey   contextKey = "user_role"
)

// UserIDFromCtx extracts user_id from context (set by UserHeaderInterceptor).
func UserIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

// UserRoleFromCtx extracts user_role from context.
func UserRoleFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(roleKey).(string)
	return v
}

// IsAdminCtx returns true if the user role is admin.
func IsAdminCtx(ctx context.Context) bool {
	return UserRoleFromCtx(ctx) == "admin"
}

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func withUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}
