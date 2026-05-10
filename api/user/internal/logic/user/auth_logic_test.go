package user

import (
	"context"
	"database/sql"
	"net/http"
	"regexp"
	"testing"
	"time"

	"GoZero-AI/api/user/internal/auth"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/statuserr"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestRegisterCreatesUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`select id,username,password_hash,created_at from "public"."users" where username = $1 limit 1`)).
		WithArgs("newuser").
		WillReturnRows(sqlmock.NewRows(authUserColumns()))
	mock.ExpectExec(regexp.QuoteMeta(`insert into "public"."users" (username,password_hash) values ($1, $2)`)).
		WithArgs("newuser", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	logic := NewRegisterLogic(context.Background(), authLogicSvcCtx(db, nil))
	resp, err := logic.Register(&types.RegisterReq{
		Username:        " newuser ",
		Password:        "secret1",
		ConfirmPassword: "secret1",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp == nil {
		t.Fatal("Register() resp = nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegisterRejectsPasswordMismatchWithBadRequest(t *testing.T) {
	logic := NewRegisterLogic(context.Background(), &svc.ServiceContext{})
	_, err := logic.Register(&types.RegisterReq{
		Username:        "newuser",
		Password:        "secret1",
		ConfirmPassword: "secret2",
	})

	assertStatusError(t, err, http.StatusBadRequest, "两次输入的密码不一致")
}

func TestRegisterDuplicateUsernameReturnsConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	passwordHash, err := auth.HashPassword("secret1")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`select id,username,password_hash,created_at from "public"."users" where username = $1 limit 1`)).
		WithArgs("newuser").
		WillReturnRows(sqlmock.NewRows(authUserColumns()).
			AddRow(int64(7), "newuser", passwordHash, time.Now()))

	logic := NewRegisterLogic(context.Background(), authLogicSvcCtx(db, nil))
	_, err = logic.Register(&types.RegisterReq{
		Username:        "newuser",
		Password:        "secret1",
		ConfirmPassword: "secret1",
	})

	assertStatusError(t, err, http.StatusConflict, "用户名已存在")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestLoginRejectsInvalidPasswordWithUnauthorized(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	passwordHash, err := auth.HashPassword("secret1")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`select id,username,password_hash,created_at from "public"."users" where username = $1 limit 1`)).
		WithArgs("alice1").
		WillReturnRows(sqlmock.NewRows(authUserColumns()).
			AddRow(int64(7), "alice1", passwordHash, time.Now()))

	logic := NewLoginLogic(context.Background(), authLogicSvcCtx(db, nil))
	_, err = logic.Login(&types.LoginReq{
		Username: "alice1",
		Password: "wrong1",
	})

	assertStatusError(t, err, http.StatusUnauthorized, "用户名或密码错误")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestLoginReturnsTokensAndStoresRefreshToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	redisClient := newMiniRedisClient(t)
	defer redisClient.Close()

	passwordHash, err := auth.HashPassword("secret1")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`select id,username,password_hash,created_at from "public"."users" where username = $1 limit 1`)).
		WithArgs("alice1").
		WillReturnRows(sqlmock.NewRows(authUserColumns()).
			AddRow(int64(7), "alice1", passwordHash, time.Now()))

	logic := NewLoginLogic(context.Background(), authLogicSvcCtx(db, redisClient))
	resp, err := logic.Login(&types.LoginReq{
		Username: "alice1",
		Password: "secret1",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.ExpireIn <= 0 {
		t.Fatalf("Login() resp = %+v, want token pair", resp)
	}

	claims, err := auth.ParseTokenWithType("auth-secret", resp.RefreshToken, auth.TokenTypeRefresh)
	if err != nil {
		t.Fatalf("ParseTokenWithType(refresh) error = %v", err)
	}
	storedUserID, err := redisClient.Get(context.Background(), auth.RefreshTokenKey(claims.ID)).Result()
	if err != nil {
		t.Fatalf("refresh token not stored in redis: %v", err)
	}
	if storedUserID != "7" {
		t.Fatalf("stored user id = %q, want 7", storedUserID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func authLogicSvcCtx(db *sql.DB, redisClient *redis.Client) *svc.ServiceContext {
	conn := sqlx.NewSqlConnFromDB(db)
	svcCtx := &svc.ServiceContext{
		DB:              conn,
		UsersModel:      model.NewUsersModel(conn),
		RedisClient:     redisClient,
		RefreshTokenTTL: time.Hour,
	}
	svcCtx.Config.Auth.AccessSecret = "auth-secret"
	svcCtx.Config.Auth.AccessExpire = 3600
	return svcCtx
}

func authUserColumns() []string {
	return []string{"id", "username", "password_hash", "created_at"}
}

func assertStatusError(t *testing.T, err error, wantCode int, wantMessage string) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %d %q", wantCode, wantMessage)
	}
	code, ok := statuserr.StatusCode(err)
	if !ok || code != wantCode {
		t.Fatalf("status = %d, ok=%v, want %d/true; err=%v", code, ok, wantCode, err)
	}
	if err.Error() != wantMessage {
		t.Fatalf("message = %q, want %q", err.Error(), wantMessage)
	}
}
