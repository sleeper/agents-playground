package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Authenticator validates inbound JWT tokens issued by the identity provider.
type Authenticator struct {
	SharedSecret []byte
}

// SubjectFromRequest extracts the subject claim from a bearer token on the request.
func (a *Authenticator) SubjectFromRequest(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("authorization header must be bearer token")
	}

	token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.SharedSecret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid claims")
	}

	subject, _ := claims["sub"].(string)
	if subject == "" {
		return "", errors.New("subject claim missing")
	}

	return subject, nil
}
