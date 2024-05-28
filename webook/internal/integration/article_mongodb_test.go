package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/wsqigo/basic-go/webook/internal/integration/startup"
	"github.com/wsqigo/basic-go/webook/internal/repository/dao"
	jwt2 "github.com/wsqigo/basic-go/webook/internal/web/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type ArticleMongoDBHandlerSuite struct {
	suite.Suite
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
	server  *gin.Engine
}

func (s *ArticleMongoDBHandlerSuite) SetupSuite() {
	s.mdb = startup.InitMongoDB()
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	hdl := startup.InitArticleHandler(dao.NewArticleMongoDBDAO(s.mdb, node))
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", jwt2.UserClaims{
			Uid: 123,
		})
	})
	hdl.RegisterRoutes(server)
	s.server = server
}

func (s *ArticleMongoDBHandlerSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.liveCol.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *ArticleMongoDBHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)

		// 前端传过来，肯定是一个 JSON
		// 构造数据，直接使用 req
		// 也就是说，我们放弃测试 Bind 的异常分支
		art Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {
				// 你要验证，保存到了数据库里面
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 验证一下数据
				var art dao.Article
				err := s.col.FindOne(ctx,
					bson.D{bson.E{Key: "author_id", Value: 123}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				art.Ctime = 0
				art.Utime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					Status:   1,
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				// 我希望你的 ID 是 1
				Data: 1,
			},
		},
		{
			name: "修改帖子",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					// 假设这是一个已经发表了的帖子
					Status:   2,
					Ctime:    456,
					Utime:    789,
					AuthorId: 123,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).
					Decode(&art)
				assert.NoError(t, err)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 789)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    456,
					// 更新之后，是未发表状态
					Status: 1,
				}, art)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)

			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "修改帖子 - 修改别人的帖子",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:      3,
					Title:   "我的标题",
					Content: "我的内容",
					// 假设这是一个已经发表了的帖子
					Status: 2,
					Ctime:  456,
					Utime:  789,
					// 模拟别人
					AuthorId: 234,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Ctime:    456,
					Utime:    789,
					Status:   2,
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Msg: "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			// 准备 Req 和记录的 recorder
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit",
				bytes.NewReader(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// 执行
			s.server.ServeHTTP(recorder, req)
			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			if tc.wantResult.Data > 0 {
				// 你只能断定有 ID
				assert.True(t, res.Data > 0)
			}
		})
	}
}

func (s *ArticleMongoDBHandlerSuite) TestPublish() {
	t := s.T()
	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)

		// 前端传过来，肯定是一个 JSON
		// 构造数据，直接使用 req
		// 也就是说，我们放弃测试 Bind 的异常分支
		art Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
				// 你要验证，保存到了数据库里面

			},
			after: func(t *testing.T) {
				// 验证一下数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 验证一下数据
				var art dao.Article
				err := s.col.FindOne(ctx,
					bson.D{bson.E{Key: "author_id", Value: 123}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, "hello, 你好", art.Title)
				assert.Equal(t, "随便试试", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, uint8(2), art.Status)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				var publishedArt dao.PublishedArticle
				err = s.liveCol.FindOne(ctx,
					bson.D{bson.E{Key: "author_id", Value: 123}}).
					Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, "hello, 你好", publishedArt.Title)
				assert.Equal(t, "随便试试", publishedArt.Content)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
				assert.Equal(t, uint8(2), publishedArt.Status)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			art: Article{
				Title:   "hello, 你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				// 我希望你的 ID 是 1
				Data: 1,
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					// 假设这是一个已经发表了的帖子
					Status:   1,
					Ctime:    456,
					Utime:    789,
					AuthorId: 123,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, uint8(2), art.Status)
				assert.Equal(t, int64(123), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 789)
				var publishedArt dao.Article
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).
					Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.True(t, publishedArt.Ctime > 0)
				assert.Equal(t, uint8(2), publishedArt.Status)
				assert.True(t, publishedArt.Utime > 0)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				art := dao.Article{
					Id:      3,
					Title:   "我的标题",
					Content: "我的内容",
					Status:  2,
					Ctime:   456,
					Utime:   789,
					// 注意。这个 AuthorID 我们设置为另一个
					AuthorId: 123,
				}
				// 模拟已经存在的帖子
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				pubArt := dao.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &pubArt)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, uint8(2), art.Status)
				assert.Equal(t, int64(123), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 789)

				var publishedArt dao.Article
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).
					Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", publishedArt.Title)
				assert.Equal(t, "新的内容", publishedArt.Content)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), publishedArt.Ctime)
				assert.Equal(t, uint8(2), publishedArt.Status)
				assert.True(t, publishedArt.Ctime > 0)
				// 更新时间变了
				assert.True(t, publishedArt.Utime > 789)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，发表失败",
			before: func(t *testing.T) {
				art := dao.Article{
					Id:      4,
					Title:   "我的标题",
					Content: "我的内容",
					Status:  1,
					Ctime:   456,
					Utime:   789,
					// 注意。这个 AuthorID 我们设置为另一个
					AuthorId: 234,
				}
				// 模拟已经存在的帖子
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				pubArt := dao.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &pubArt)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, uint8(1), art.Status)
				assert.Equal(t, int64(234), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间没变
				assert.Equal(t, int64(789), art.Utime)

				var publishedArt dao.Article
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).
					Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, uint8(1), art.Status)
				assert.Equal(t, int64(234), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间没变
				assert.Equal(t, int64(789), art.Utime)
			},
			art: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			// 准备 Req 和记录的 recorder
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish",
				bytes.NewReader(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// 执行
			s.server.ServeHTTP(recorder, req)
			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.NoError(t, err)
			if tc.wantResult.Data > 0 {
				// 你只能断定有 ID
				assert.True(t, res.Data > 0)
			}
		})
	}
}

func TestArticleMongoDBHandler(t *testing.T) {
	suite.Run(t, &ArticleMongoDBHandlerSuite{})
}
