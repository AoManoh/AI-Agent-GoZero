package user

import (
	"context"
	"encoding/json"
	"math"
	"strconv"

	"GoZero-AI/internal/statuserr"
)

func currentUserID(ctx context.Context) (int64, error) {
	raw := ctx.Value("userId")
	if raw == nil {
		return 0, statuserr.Unauthorized("未找到当前用户身份")
	}

	switch value := raw.(type) {
	case int64:
		return normalizeUserID(value)
	case int32:
		return normalizeUserID(int64(value))
	case int:
		return normalizeUserID(int64(value))
	case float64:
		if value < 0 || math.Trunc(value) != value {
			return 0, statuserr.Unauthorized("当前用户身份无效")
		}
		return normalizeUserID(int64(value))
	case json.Number:
		parsed, err := value.Int64()
		if err != nil {
			return 0, statuserr.Unauthorized("当前用户身份无效")
		}
		return normalizeUserID(parsed)
	case string:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, statuserr.Unauthorized("当前用户身份无效")
		}
		return normalizeUserID(parsed)
	default:
		return 0, statuserr.Unauthorized("当前用户身份无效")
	}
}

func normalizeUserID(userID int64) (int64, error) {
	if userID <= 0 {
		return 0, statuserr.Unauthorized("当前用户身份无效")
	}
	return userID, nil
}
