package web

import "github.com/gin-gonic/gin"

type OAuth2WechatHandler struct {
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("")
	g.Any("")
}
