package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/wsqigo/basic-go/webook/internal/service"
	"github.com/wsqigo/basic-go/webook/internal/service/oauth2/dingding"
	jwt2 "github.com/wsqigo/basic-go/webook/internal/web/jwt"
	"github.com/wsqigo/basic-go/webook/pkg/ginx"
	"net/http"
)

type OAuth2DingDingHandler struct {
	jwt2.Handler
	svc             dingding.Service
	userSvc         service.UserService
	key             []byte
	stateCookieName string
}

func NewOAuth2DingDingHandler(svc dingding.Service,
	hdl jwt2.Handler, userSvc service.UserService) *OAuth2DingDingHandler {
	return &OAuth2DingDingHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("S4EWBerIvPWZDfH9jpFRBByIE5HcBfiP"),
		stateCookieName: "jwt-state",
		Handler:         hdl,
	}
}

func (o *OAuth2DingDingHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/dingding")
	g.GET("/authurl", o.OAuth2URL)
	// 这边用 Any 万无一失
	g.Any("/callback", o.Callback)
}

func (o *OAuth2DingDingHandler) OAuth2URL(ctx *gin.Context) {
	state := uuid.New()
	val, err := o.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
		return
	}
	err = o.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "服务器异常",
			Code: 5,
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: val,
	})
}

func (o *OAuth2DingDingHandler) Callback(ctx *gin.Context) {
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}
	// 你校验不校验都可以
	code := ctx.Query("code")
	//state := ctx.Query("state")
	dDingInfo, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "授权码有误",
			Code: 4,
		})
		return
	}

	u, err := o.userSvc.FindOrCreateByDDing(ctx, dDingInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}

	err = o.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}

func (o *OAuth2DingDingHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie(o.stateCookieName)
	if err != nil {
		return fmt.Errorf("无法获得 cookie %w", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}
	if state != sc.State {
		// state 不匹配，有人搞你
		return fmt.Errorf("state 不匹配")
	}
	return nil
}

// setStateJWT 只有钉钉这里用，所以定义在这里
func (o *OAuth2DingDingHandler) setStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(o.key)
	if err != nil {
		return err
	}
	ctx.SetCookie(o.stateCookieName, tokenStr,
		// 限制在只能在这里生效。
		600, "/oauth2/dingding/callback",
		// 这边把 HTTPS 协议禁止了，不过在生产环境中要开启。
		"", true, true)
	return nil
}
