package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	jwt2 "github.com/wsqigo/basic-go/webook/internal/web/jwt"
	"net/http"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string

	jwt2.Handler
}

func NewLoginJWTMiddlewareBuilder(hdl jwt2.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: hdl,
	}
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		//if ctx.Request.URL.Path == "/users/login" ||
		//	ctx.Request.URL.Path == "/user/signup" {
		//	return
		//}
		for _, path := range m.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 根据约定，token 在 Authorization 这个头部
		// 得到的格式 Bearer token
		tokenStr := m.ExtractToken(ctx)
		var uc jwt2.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return jwt2.JWTKey, nil
		})
		if err != nil {
			// token 不对，token 是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			// 在这里发现 access_token 过期了，生成一个新的 access_token
			// 相当于自动刷新

			// token 解析出来了，但是 token 可能是非法的，或者过期了的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 这里看
		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// token 无效或者 redis 有问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 可以兼容 Redis 异常的情况
		// 做好监控，监控有没有 error
		//if cnt > 0 {
		//	// token 无效或者 redis 有问题
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		ctx.Set("user", uc)
	}
}

func (m *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	m.paths = append(m.paths, path)
	return m
}
