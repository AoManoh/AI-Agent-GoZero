package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken     string
	RefreshToken    string
	ExpireIn        int64
	RefreshTokenJTI string
}

func IssueTokenPair(secret string, accessTTL, refreshTTL time.Duration, userID int64, username string) (*TokenPair, error) {
	accessToken, _, err := issueToken(secret, accessTTL, userID, username, TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshJTI, err := issueToken(secret, refreshTTL, userID, username, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		ExpireIn:        int64(accessTTL / time.Second),
		RefreshTokenJTI: refreshJTI,
	}, nil
}

func ParseToken(secret, token string) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok || !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func ParseTokenWithType(secret, token, expectedType string) (*Claims, error) {
	claims, err := ParseToken(secret, token)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("unexpected token type: %s", claims.TokenType)
	}

	return claims, nil
}

func RefreshTokenKey(jti string) string {
	return "user:refresh:" + jti
}

func UserRefreshIndexKey(userID int64) string {
	return "user:refresh:index:" + strconv.FormatInt(userID, 10)
}

func issueToken(secret string, ttl time.Duration, userID int64, username, tokenType string) (string, string, error) {
	now := time.Now()
	jti := uuid.NewString()
	claims := Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	return signed, jti, nil
}
