package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/maxkulish/service-api/business/sys/auth"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestAuth(t *testing.T) {
	t.Log("Given the need to be able to authenticate and authorize access.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single user.", testID)
		{
			const keyID = "59189175-db33-438d-8cc7-4abde89fbbe8"
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a private key: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a private key.", success, testID)

			a, err := auth.NewAuth(keyID, &keyStore{pk: privateKey})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create an authenticator: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create an authenticator.", success, testID)

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "test",
					Subject:   "59189175-db33-438d-8cc7-4abde89fbbe8",
					ExpiresAt: jwt.NewNumericDate(jwt.TimeFunc().Add(1 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(jwt.TimeFunc()),
				},
				Roles: []string{auth.RoleAdmin},
			}

			token, err := a.GenerateToken(claims)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a JWT token: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a JWT token.", success, testID)

			parsedClaims, err := a.ValidateToken(token)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse the claims: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse the claims.", success, testID)

			if exp, got := len(claims.Roles), len(parsedClaims.Roles); exp != got {
				t.Logf("\t\tTest %d:\tExpected: %d", testID, exp)
				t.Logf("\t\tTest %d:\tGot: %d", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the correct number of roles: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the correct number of roles.", success, testID)

			if exp, got := claims.Roles[0], parsedClaims.Roles[0]; exp != got {
				t.Logf("\t\tTest %d:\tExpected: %v", testID, exp)
				t.Logf("\t\tTest %d:\tGot: %v", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the correct number of roles: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the correct number of roles.", success, testID)
		}
	}
}

// =============================================================================

type keyStore struct {
	pk *rsa.PrivateKey
}

func (ks *keyStore) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	return ks.pk, nil
}

func (ks *keyStore) PublicKey(kid string) (*rsa.PublicKey, error) {
	return &ks.pk.PublicKey, nil
}
