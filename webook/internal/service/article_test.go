package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/repository"
	repomocks "github.com/wsqigo/basic-go/webook/internal/repository/mocks"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (
			repository.ArticleAuthorRepository,
			repository.ArticleReaderRepository)

		art domain.Article

		wantId  int64
		wantErr error
	}{
		{
			name: "发表成功",
			mock: func(ctrl *gomock.Controller) (
				repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Return(int64(1), nil)
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用了制作库的 ID
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				})
				return authorRepo, readerRepo
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
				Status: 2,
			},
			wantId: 1,
		},
		{
			name: "修改并保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (
				repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Return(errors.New("mock db error"))
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				return authorRepo, readerRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
				Status: 2,
			},
			wantErr: errors.New("mock db error"),
		},
		{
			name: "修改并新发表成功",
			mock: func(ctrl *gomock.Controller) (
				repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Return(nil)
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用了制作库的 ID
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Return(nil)
				return authorRepo, readerRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
				Status: 2,
			},
			wantId: 11,
		},
		{
			name: "修改成功但发表失败-重试成功",
			mock: func(ctrl *gomock.Controller) (
				repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Return(nil)
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用了制作库的 ID
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Return(errors.New("mock db error"))
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用了制作库的 ID
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Return(nil)
				return authorRepo, readerRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
				Status: 2,
			},
			wantId:  11,
			wantErr: nil,
		},
		{
			name: "修改成功但发表失败-重试失败",
			mock: func(ctrl *gomock.Controller) (
				repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				authorRepo := repomocks.NewMockArticleAuthorRepository(ctrl)
				authorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Return(nil)
				readerRepo := repomocks.NewMockArticleReaderRepository(ctrl)
				readerRepo.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用了制作库的 ID
					Id:      11,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
					Status: 2,
				}).Times(3).Return(errors.New("mock db error"))
				return authorRepo, readerRepo
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
				Status: 2,
			},
			wantId:  11,
			wantErr: errors.New("保存到线上库失败，重试次数耗尽"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo, readerRepo := tc.mock(ctrl)
			svc := NewArticleServiceV1(authorRepo, readerRepo, logger.NewNopLogger())
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
