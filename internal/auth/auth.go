package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"kubernetes-api/internal/models"
	"kubernetes-api/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

var (
	jwtKey []byte
)

// Claims represents the JWT claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// InitAuth initializes the authentication module
func InitAuth() error {
	// Use environment variable or generate a random key
	keyStr := os.Getenv("JWT_SECRET")
	if keyStr == "" {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			return fmt.Errorf("failed to generate JWT key: %w", err)
		}
		keyStr = base64.StdEncoding.EncodeToString(key)
		logrus.Warn("JWT_SECRET not set, generated random key. This is not suitable for production.")
	}
	jwtKey = []byte(keyStr)
	return nil
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a bcrypt hashed password with its plaintext version
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kubernetes-api",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ExtractTokenFromRequest extracts JWT token from Authorization header
func ExtractTokenFromRequest(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

// AuthMiddleware is a middleware to authenticate requests
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := ExtractTokenFromRequest(r)
		if tokenString == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context using custom keys
		ctx := r.Context()
		ctx = context.WithValue(ctx, utils.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, utils.UsernameKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
