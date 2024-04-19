package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Time{})
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		//if ctx.Request.URL.Path == "/users/login" ||
		//	ctx.Request.URL.Path == "/user/signup" {
		//	return
		//}
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			// 没有登录
			// 中断，不要往后执行，也就是不要执行后面的业务逻辑
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()

		// 我怎么知道，要刷新了呢？
		// 假如说，我们的策略是每分钟刷一次，我怎么知道，已经过了一分钟？
		const updateTimeKey = "update_time"
		// 试着拿出上一次刷新时间
		updateTime := sess.Get(updateTimeKey)
		lastUpdateTime, ok := updateTime.(time.Time)
		if updateTime == nil || !ok || now.Sub(lastUpdateTime) > time.Minute {
			// 说明还没有刷新过，刚登陆，还没刷新过
			sess.Set(updateTimeKey, now)
			sess.Set("userId", userId)
			err := sess.Save()
			if err != nil {
				// 打日志
				fmt.Println(err)
			}
			return
		}
	}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}
