package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/wsqigo/basic-go/webook/internal/repository/cache/redismocks"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRedisUserCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(0))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "15212345678")}, []any{"123456"}).
					Return(cmd)
				return res
			},
			ctx:   context.Background(),
			biz:   "test",
			phone: "15212345678",
			code:  "123456",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewCodeCache(tc.mock(ctrl))
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
