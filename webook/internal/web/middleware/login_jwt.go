package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wsqigo/basic-go/webook/internal/web"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
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
		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			// 没登录，没有 token，Authorization 这个头部都没有
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(authCode, " ")
		if len(segs) != 2 {
			// 没登录，Authorization 中的内容是乱传的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		if err != nil {
			// token 不对，token 是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			// token 解析出来了，但是 token 可能是非法的，或者过期了的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//if uc.UserAgent != ctx.GetHeader("User-Agent") {
		//	// 后期我们讲到了监控告警的时候，这个地方要埋点
		//	// 能够进来这个分支的，大概率是攻击者
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		expireTime := uc.ExpiresAt
		// 不判定都可以
		//if expireTime.Before(time.Now()) {
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		// 剩余过期时间 < 50s 就要刷新
		if expireTime.Sub(time.Now()) < 50*time.Second {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(30 * time.Minute))
			tokenStr, err = token.SignedString(web.JWTKey)
			ctx.Header("x-jwt-token", tokenStr)
			if err != nil {
				// 这边不要中断，因为仅仅是过期时间没有刷新，但是用户是登录了的
				log.Println(err)
			}
		}
		ctx.Set("user", uc)
	}
}

func (m *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	m.paths = append(m.paths, path)
	return m
}
