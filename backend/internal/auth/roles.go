package auth

import (
	"net/http"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

func RequireRole(allowed ...Role) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userCtx := r.Context().Value(UserContextKey)

			if userCtx == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, ok := userCtx.(*UserClaims)
			if !ok {
				http.Error(w, "invalid user context", http.StatusUnauthorized)
				return
			}

			if hasRole(user.Roles, allowed) {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

func hasRole(userRoles []Role, allowed []Role) bool {

	for _, ur := range userRoles {
		for _, ar := range allowed {
			if ur == ar {
				return true
			}
		}
	}

	return false
}
