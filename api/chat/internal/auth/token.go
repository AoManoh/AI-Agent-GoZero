package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

const tokenTypeAccess = "access"

type userIDContextKey struct{}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(int64)
	return userID, ok
}

func ParseAccessTokenUserID(secret, token string) (int64, error) {
	if secret == "" {
		return 0, errors.New("chat service 未配置 access secret")
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	if !parsedToken.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}
	if tokenType, ok := claims["tokenType"].(string); ok && tokenType != "" && tokenType != tokenTypeAccess {
		return 0, errors.New("unexpected token type")
	}

	rawUserID, ok := claims["userId"]
	if !ok {
		return 0, errors.New("token missing userId")
	}

	return parseUserID(rawUserID)
}

func parseUserID(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unknown userId type: %T", value)
	}
}
