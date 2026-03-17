package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type Auth struct {
	verifier *oidc.IDTokenVerifier
}

type UserClaims struct {
	Username string `json:"preferred_username"`
	Email    string `json:"email"`
	Roles    []Role `json:"roles"`
}

type contextKey string

const UserContextKey contextKey = "user"

func NewAuth(issuer string) (*Auth, error) {

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true,
	})

	return &Auth{
		verifier: verifier,
	}, nil
}

func (a *Auth) Middleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString, err := extractBearerToken(r)
		if err != nil {
			http.Error(w, "missing authorization token", http.StatusUnauthorized)
			return
		}

		idToken, err := a.verifier.Verify(r.Context(), tokenString)
		if err != nil {
			log.Println("token verify error:", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		var rawClaims map[string]interface{}
		if err := idToken.Claims(&rawClaims); err != nil {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		user := parseClaims(rawClaims)

		ctx := context.WithValue(r.Context(), UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1], nil
		}
	}

	if token := r.URL.Query().Get("token"); token != "" {
		return token, nil
	}

	return "", errors.New("missing authorization token")
}

func parseClaims(claims map[string]interface{}) *UserClaims {

	user := &UserClaims{}

	if username, ok := claims["preferred_username"].(string); ok {
		user.Username = username
	}

	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}

	if resourceAccess, ok := claims["resource_access"].(map[string]interface{}); ok {

		if clientRoles, ok := resourceAccess["lan-control-plane"].(map[string]interface{}); ok {

			if roles, ok := clientRoles["roles"].([]interface{}); ok {

				for _, r := range roles {
					if role, ok := r.(string); ok {
						user.Roles = append(user.Roles, Role(role))
					}
				}

			}
		}
	}

	return user
}
