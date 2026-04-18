package user

import (
	"context"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"

	"GoZero-AI/api/user/internal/auth"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/internal/statuserr"
)

type accessTokenContextKey struct{}

func WithAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, accessTokenContextKey{}, token)
}

func accessTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(accessTokenContextKey{}).(string)
	return token
}

func storeRefreshToken(ctx context.Context, svcCtx *svc.ServiceContext, userID int64, refreshTokenJTI string) error {
	userIDValue := strconv.FormatInt(userID, 10)
	indexKey := auth.UserRefreshIndexKey(userID)
	refreshKey := auth.RefreshTokenKey(refreshTokenJTI)

	pipe := svcCtx.RedisClient.TxPipeline()
	pipe.Set(ctx, refreshKey, userIDValue, svcCtx.RefreshTokenTTL)
	pipe.SAdd(ctx, indexKey, refreshTokenJTI)
	pipe.Expire(ctx, indexKey, svcCtx.RefreshTokenTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return redisStorageUnavailable("写入 refresh token", err)
	}
	return nil
}

func validateRefreshToken(ctx context.Context, svcCtx *svc.ServiceContext, userID int64, refreshTokenJTI string) error {
	userIDValue, err := svcCtx.RedisClient.Get(ctx, auth.RefreshTokenKey(refreshTokenJTI)).Result()
	if errors.Is(err, redis.Nil) {
		return errors.New("refresh token 已失效，请重新登录")
	}
	if err != nil {
		return redisStorageUnavailable("校验 refresh token", err)
	}
	if userIDValue != strconv.FormatInt(userID, 10) {
		return errors.New("refresh token 不属于当前用户")
	}

	return nil
}

func rotateRefreshToken(ctx context.Context, svcCtx *svc.ServiceContext, userID int64, oldRefreshTokenJTI, newRefreshTokenJTI string) error {
	userIDValue := strconv.FormatInt(userID, 10)
	indexKey := auth.UserRefreshIndexKey(userID)

	pipe := svcCtx.RedisClient.TxPipeline()
	pipe.Set(ctx, auth.RefreshTokenKey(newRefreshTokenJTI), userIDValue, svcCtx.RefreshTokenTTL)
	pipe.SAdd(ctx, indexKey, newRefreshTokenJTI)
	pipe.Expire(ctx, indexKey, svcCtx.RefreshTokenTTL)
	pipe.Del(ctx, auth.RefreshTokenKey(oldRefreshTokenJTI))
	pipe.SRem(ctx, indexKey, oldRefreshTokenJTI)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return redisStorageUnavailable("轮换 refresh token", err)
	}
	return nil
}

func revokeAllRefreshTokens(ctx context.Context, svcCtx *svc.ServiceContext, userID int64) error {
	indexKey := auth.UserRefreshIndexKey(userID)
	tokenJTIs, err := svcCtx.RedisClient.SMembers(ctx, indexKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return redisStorageUnavailable("读取 refresh token 索引", err)
	}

	pipe := svcCtx.RedisClient.TxPipeline()
	for _, jti := range tokenJTIs {
		pipe.Del(ctx, auth.RefreshTokenKey(jti))
	}
	pipe.Del(ctx, indexKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return redisStorageUnavailable("撤销 refresh token", err)
	}
	return nil
}

func redisStorageUnavailable(_ string, err error) error {
	if err == nil {
		return nil
	}
	return statuserr.ServiceUnavailable("认证会话存储不可用，请稍后重试")
}
