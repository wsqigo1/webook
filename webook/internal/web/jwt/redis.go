package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisJWTHandler struct {
	client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration
}

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		client:        client,
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  7 * 24 * time.Hour,
	}
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.client.Exists(ctx, h.key(ssid)).Result()
	if err != nil {
		return err
	}

	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	// 正常用户的这两个 token 都会被前端更新
	// 也就是说在登录校验里面，走不到 redis 那一步就返回了
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	// 这里不可能拿不到
	uc := ctx.MustGet("user").(UserClaims)
	return h.client.Set(ctx, h.key(uc.Ssid), "", h.rcExpiration).Err()
}

// ExtractToken 根据约定，token 在 Authorization 这个头部
// 得到的格式 Bearer token
func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return authCode
	}

	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

var _ Handler = &RedisJWTHandler{}

func (h *RedisJWTHandler) key(ssid string) string {
	return fmt.Sprintf("users:ssid:%s", ssid)
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}

	return h.SetJWTToken(ctx, uid, ssid)
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			// 30 分钟过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置为7天过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
	}
	refreshToken := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := refreshToken.SignedString(RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

var JWTKey = []byte("S4EWBerIvPWZDfH8jpFRBByIE5HcBfiP")
var RCJWTKey = []byte("S4EWBerIvPWZDfH8jpFRBByIE5HdBfiP")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}
