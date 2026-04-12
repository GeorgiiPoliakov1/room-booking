package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTService struct {
	secret string
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

var (
	ErrInvalidToken = errors.New("invalid or malformed token")
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidRole  = errors.New("invalid role in token")
)

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: secret,
	}
}

func (s *JWTService) GenerateToken(userID uuid.UUID, role string) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrInvalidToken
		}
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrExpiredToken
			}
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrInvalidToken
			}
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (c *Claims) ParseUserID() (uuid.UUID, error) {
	if c.UserID == "" {
		return uuid.Nil, errors.New("empty user_id in claims")
	}
	return uuid.Parse(c.UserID)
}
