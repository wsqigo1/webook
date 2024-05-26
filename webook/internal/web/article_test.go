package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/service"
	svcmocks "github.com/wsqigo/basic-go/webook/internal/service/mocks"
	jwt2 "github.com/wsqigo/basic-go/webook/internal/web/jwt"
	"github.com/wsqigo/basic-go/webook/pkg/ginx"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock    func(ctrl *gomock.Controller) service.ArticleService
		reqBody string

		wantCode int
		wantRes  ginx.Result
	}{
		{
			name: "新建并发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"title": "我的标题",
	"content": "我的内容"
}
`,
			wantCode: 200,
			wantRes: ginx.Result{
				Data: float64(1),
			},
		},
		{
			name: "修改并且发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"id": 1,
	"title": "我的标题",
	"content": "我的内容"
}
`,
			wantCode: 200,
			wantRes: ginx.Result{
				// 原本是 int64的，但是因为 Data 是any，所以在反序列化的时候，
				// 用的 float64
				Data: float64(1),
			},
		},
		{
			name: "输入有误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				return svc
			},
			reqBody: `
{
	"id": 1,
	"title": "我的标题",
	"content": "我的内容",
}
`,
			wantCode: 400,
		},
		{
			name: "publish错误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("mock error"))
				return svc
			},
			reqBody: `
{
	"id": 1,
	"title": "我的标题",
	"content": "我的内容"
}
`,
			wantCode: 200,
			wantRes: ginx.Result{
				Msg:  "系统错误",
				Code: 5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 构造 handler
			svc := tc.mock(ctrl)
			hdl := NewArticleHandler(svc, logger.NewNopLogger())

			// 准备服务器，注册路由
			server := gin.Default()
			// 设置登录态
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", jwt2.UserClaims{
					Uid: 123,
				})
			})
			hdl.RegisterRoutes(server)

			// 准备 Req 和记录的 recorder
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish",
				bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			// 准备记录响应
			recorder := httptest.NewRecorder()

			// 执行
			server.ServeHTTP(recorder, req)
			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}

			var res ginx.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
