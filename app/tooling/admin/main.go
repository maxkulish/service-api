package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func main() {

	err := genToken()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func genToken() error {

	// Read the private key from the file
	privateFileName := "zarf/keys/59189175-db33-438d-8cc7-4abde89fbbe8.pem"
	file, err := os.Open(privateFileName)
	if err != nil {
		return fmt.Errorf("reading auth private key file: %w", err)
	}
	defer file.Close()

	// limit PEM file size to 1MB. This should be reasonable for almost
	// any PEM file and prevents shenanigans like linking the file
	// to /dev/random or something like that.
	privatePEM, err := io.ReadAll(io.LimitReader(file, 1_024*1_024))
	if err != nil {
		return fmt.Errorf("reading auth private key file: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return fmt.Errorf("parsing auth private key file: %w", err)
	}

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service-api project",
			Subject:   "123456789",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod("RS256")
	token := jwt.NewWithClaims(method, claims)
	// We need to rotate the key used to sign the token once a month.
	// This is done by adding a key ID to the header. This key ID
	// will be used to lookup the correct key to verify the token.
	token.Header["kid"] = "59189175-db33-438d-8cc7-4abde89fbbe8" // Key ID

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	fmt.Println("============= TOKEN BEGIN =============")
	fmt.Println(tokenStr)
	fmt.Println("============= TOKEN END =============")
	fmt.Println()

	// =========================================================================

	// Marshal the public key for the private key to PKIX
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Create a file for the public key information in PEM format
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Construct a PEM block for the public key
	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the private key file
	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	// =========================================================================

	// Create the token parser to use. The algorithm used to sign the JWT
	// must be validated to avoid a critical vulnerability.
	// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := jwt.Parser{
		ValidMethods: []string{"RS256"},
	}

	var parsedClaims struct {
		jwt.RegisteredClaims
		Roles []string
	}

	keyFunc := func(t *jwt.Token) (any, error) {
		kind, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("kid not found in token header")
		}
		kindID, ok := kind.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be a string")
		}
		fmt.Println("========================")
		fmt.Println("KID:", kindID)
		return &privateKey.PublicKey, nil
	}

	parsedToken, err := parser.ParseWithClaims(tokenStr, &parsedClaims, keyFunc)
	if err != nil {
		return fmt.Errorf("parsing token: %w", err)
	}

	if !parsedToken.Valid {
		return errors.New("token is not valid")
	}

	fmt.Println("========================")
	fmt.Println("Token validated")

	return nil
}

// genKey creates an x509 private/public key for auth tokens
func genKey() error {

	// Generate a new private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// Create a file for the private key information in PEM format
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key
	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	// =========================================================================

	// Marshal the public key for the private key to PKIX
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Create a file for the public key information in PEM format
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Construct a PEM block for the public key
	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the private key file
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	return nil
}
