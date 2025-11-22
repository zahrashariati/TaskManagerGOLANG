// Private key: Supports PKCS#1 and PKCS#8 formats
// Public key: Uses PKIX/X.509 format (default for OpenSSL)
package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"errors"
	"strings"
	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
	apperrors "task_manager/internal/errors"
)

type JWTService struct {
	privateKey *rsa.PrivateKey
	publicKey *rsa.PublicKey
}

// - Read private key file
// - Decode PEM
// - Parse RSA private key
// - Read public key file
// - Decode PEM
// - Parse RSA public key
// - Return JWTService with both keys

func NewJWTService(privateKeyBytes, publicKeyBytes []byte) (*JWTService, error) {
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, apperrors.Wrap(errors.New("failed to decode PEM block"), "error decoding private key file")
	}
	
	// Try PKCS#1 first, then PKCS#8
	var privateKey *rsa.PrivateKey
	var err error
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// If PKCS#1 fails, try PKCS#8
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, apperrors.Wrap(err, "error parsing private key (tried both PKCS#1 and PKCS#8)")
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, apperrors.Wrap(errors.New("private key is not RSA"), "error parsing private key")
		}
	}
	
	block, _ = pem.Decode(publicKeyBytes)
	if block == nil {
		return nil, apperrors.Wrap(errors.New("failed to decode PEM block"), "error decoding public key file")
	}
	
	// Parse PKIX (X.509) format (default for openssl rsa -pubout)
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, apperrors.Wrap(err, "error parsing public key (expected PKIX/X.509 format)")
	}
	publicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, apperrors.Wrap(errors.New("public key is not RSA"), "error parsing public key")
	}
	return &JWTService{
		privateKey: privateKey,
		publicKey: publicKey,
	}, nil
}

//Claims are the data stored inside the JWT token. When you decode a token, you get these fields. RegisteredClaims includes expiration time, issued time, etc.
type Claims struct {
	UserID int `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken function:
// - Create Claims with userID, username, expiration (24 hours)
// - Create JWT token with RS256 method
// - Sign with private key
// - Return token string

func (s *JWTService) GenerateToken(userId int, username string, expiresAt time.Time) (string, error) {
	claims := &Claims{
		UserID: userId,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer: "task_manager",
			Subject: fmt.Sprintf("%d", userId),
			Audience: jwt.ClaimStrings{"task_manager"},
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID: uuid.New().String(),
		},
	}
	// Create JWT token with RS256 method
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims) //creates new token

	// Sign with private key
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", apperrors.Wrap(err, "error signing token")
	}
	return tokenString, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok { //RSA signing method
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "error parsing token")
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, apperrors.Wrap(errors.New("token claims are invalid"), "invalid token")
}


func (s *JWTService) GetTokenFromHeader(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return "", apperrors.Wrap(errors.New("missing authorization header"), "error getting token from header")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", apperrors.Wrap(errors.New("invalid authorization header"), "error getting token from header")
	}
	return parts[1], nil
}

