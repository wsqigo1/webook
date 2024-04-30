package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/service"
	"net/http"
	"time"
)

const (
	emailRegexPattern    = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	bizLogin             = "login"
)

// UserHandler 用户路由
type UserHandler struct {
	svc            *service.UserService
	codeSvc        *service.CodeService
	emailEexReg    *regexp.Regexp
	passwordRegExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService, codeSvc *service.CodeService) *UserHandler {
	return &UserHandler{
		svc:            svc,
		codeSvc:        codeSvc,
		emailEexReg:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// 分组
	ug := server.Group("/users")
	ug.GET("/profile", h.Profile)
	ug.POST("/signup", h.SignUp)
	//ug.POST("/login", h.Login)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/logout", h.Logout)
	ug.POST("/edit", h.Edit)

	// 手机验证码登录相关功能
	ug.POST("/login_sms/code/send", h.SendSMSLoginCode)
	ug.POST("/login_sms", h.LoginSMS)
}

func (h *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := h.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码不对，请重新输入",
		})
		return
	}

	u, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	h.setJWTToken(ctx, u.Id)
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "登录成功",
	})
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 你这边可以校验 Req
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "请输入手机号码",
		})
		return
	}

	err := h.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 补日志的
	}
}

// SignUp 用户注册接口
func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	// Bind 方法会根据 Content-Type 来解析你的数据到 req 里面
	// 解析错了，就会直接写回一个 400 错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailEexReg.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不一致")
		return
	}

	isPassword, err := h.passwordRegExp.MatchString(req.ConfirmPassword)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	// 调用一下 svc 方法
	err = h.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		// 在这里登录成功了
		// 设置 session
		sess := sessions.Default(ctx)
		// 我可以随意设置值了
		// 你要放在 session 里面的值
		sess.Set("userId", user.Id)
		sess.Options(sessions.Options{
			//Secure: true,
			//HttpOnly: true,
			// 15 分钟
			MaxAge: 900,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}

		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		h.setJWTToken(ctx, user.Id)
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) setJWTToken(ctx *gin.Context, uid int64) {
	uc := UserClaims{
		Uid:       uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			// 30 分钟过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
}

// Profile 用户详情
func (h *UserHandler) Profile(ctx *gin.Context) {
	us := ctx.MustGet("user").(UserClaims)
	user, err := h.svc.FindById(ctx, us.Uid)
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那说明是系统出了问题
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		AboutMe  string `json:"aboutMe"`
		Birthday string `json:"birthday"`
	}

	ctx.JSON(http.StatusOK, User{
		Email:    user.Email,
		Nickname: user.Nickname,
		AboutMe:  user.AboutMe,
		Phone:    user.Phone,
		Birthday: user.Birthday.Format(time.DateOnly),
	})
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		// 改邮箱，密码，或者能不能改手机号
		Nickname string `json:"nickname"`
		// YYYY-MM-DD
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	var req EditReq
	// Bind 方法会根据 Content-Type 来解析你的数据到 req 里面
	// 解析错了，就会直接写回一个 400 错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	uc, ok := ctx.MustGet("user").(UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 用户输入不对
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}

	// 调用一下 svc 方法
	err = h.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Uid,
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

func (h *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	sess.Save()

	ctx.String(http.StatusOK, "退出登录成功")
}

var JWTKey = []byte("S4EWBerIvPWZDfH8jpFRBByIE5HcBfiP")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
