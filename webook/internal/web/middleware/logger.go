package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type LogMiddlewareBuilder struct {
	logFn         func(ctx context.Context, al AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFn func(ctx context.Context, al AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFn: logFn,
	}
}

func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		path := ctx.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}
		al := AccessLog{
			Path:   path,
			Method: ctx.Request.Method,
		}

		if l.allowReqBody && ctx.Request.Body != nil {
			// 直接忽略 error，不影响程序运行
			reqBodyBytes, _ := ctx.GetRawData()
			if len(reqBodyBytes) <= 2048 {
				al.ReqBody = string(reqBodyBytes)
			} else {
				al.ReqBody = string(reqBodyBytes[:2048])
			}

			// Request.Body 是一个 Stream（流）对象，所以是只能读取一次的
			// 因此读完之后要放回去，不然后续步骤是读不到的
			ctx.Request.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))
			//ctx.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		if l.allowRespBody {
			ctx.Writer = &responseWriter{
				ResponseWriter: ctx.Writer,
				al:             &al,
			}
		}

		defer func() {
			al.Duration = time.Since(start)
			//if l.allowReqBody {
			//	// 我怎么拿到这个 response 里面的数据呢？
			//
			//}
			l.logFn(ctx, al)
		}()

		// 直接执行下一个 middleware...直到业务逻辑
		ctx.Next()

		// 在这里，你就拿到了响应
	}
}

// AccessLog 你可以打印很多的信息，根据需要自己加
type AccessLog struct {
	StatusCode int           `json:"status_code"`
	Path       string        `json:"path"`
	Method     string        `json:"method"`
	ReqBody    string        `json:"req_body"`
	RespBody   string        `json:"resp_body"`
	Duration   time.Duration `json:"duration"`
}

type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (r responseWriter) WriteHeader(statusCode int) {
	r.al.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r responseWriter) Write(data []byte) (int, error) {
	r.al.RespBody = string(data)
	return r.ResponseWriter.Write(data)
}

func (r responseWriter) WriteString(data string) (int, error) {
	r.al.RespBody = data
	return r.ResponseWriter.WriteString(data)
}
