package web

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/service"
	"github.com/wsqigo/basic-go/webook/internal/web/jwt"
	"github.com/wsqigo/basic-go/webook/pkg/ginx"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	l        logger.LoggerV1
	svc      service.ArticleService
	interSvc service.InteractiveService
	biz      string
}

func NewArticleHandler(l logger.LoggerV1, svc service.ArticleService,
	interSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		l:        l,
		svc:      svc,
		interSvc: interSvc,
		biz:      "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")

	//g.PUT("/", h.Edit())
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)

	// 创作者接口
	// 在有 list 等路由的时候，无法这样注册
	//g.GET(":Id", h.Detail)
	g.GET("/detail/:id", h.Detail)
	// 理论上来说应该用 GET 的，但是我实在不耐烦处理类型转化
	// 直接 POST，JSON 转一了百了
	// /list?offset=?&limit=?
	g.POST("/list", h.List)

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 传入一个参数，true 就是点赞, false 就是不点赞
	pub.POST("/like", h.Like)
	pub.POST("/collect", h.Collect)
}

// Edit 接收 Article 输入，输入一个 ID，文章的 ID
func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		// 有 id 代表更新，没有 id 代表新建
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg: "系统错误",
		})
		h.l.Error("保存文章数据失败",
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		// 有 id 代表更新，没有 id 代表新建
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("发送文章数据失败",
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	err := h.svc.Withdraw(ctx, uc.Uid, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("撤回文章失败",
			logger.Int64("uid", uc.Uid),
			logger.Int64("art_id", req.Id),
			logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "参数错误",
			Code: 4,
		})
		h.l.Warn("查询文章失败，id 格式不对",
			logger.String("id", idStr),
			logger.Error(err))
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查询文章失败",
			logger.Error(err),
			logger.Int64("id", id))
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	if art.Author.Id != uc.Uid {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("非法查询文章",
			logger.Int64("id", id),
			logger.Int64("uid", uc.Uid))
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{
		Data: ArticleVo{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract:   art.Abstract(),
			Content: art.Content,
			Status:  art.Status.ToUint8(),
			// 这个是传作者看看自己的文章列表，也不需要这个字段
			AuthorId: art.Author.Id,
			Ctime:    art.Ctime.Format(time.DateTime),
			Utime:    art.Utime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	// 我要不要检测一下
	uc := ctx.MustGet("user").(jwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:    src.Id,
				Title: src.Title,

				Content: src.Content,
				Status:  src.Status.ToUint8(),
				Ctime:   src.Ctime.Format(time.DateTime),
				Utime:   src.Utime.Format(time.DateTime),
			}
		}),
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "参数错误",
			Code: 4,
		})
		h.l.Warn("查询文章失败，id 格式不对",
			logger.String("id", idStr),
			logger.Error(err))
		return
	}

	// 使用 error group 来同时查询
	var (
		eg    errgroup.Group
		art   domain.Article
		inter domain.Interactive
	)

	eg.Go(func() error {
		var er error
		art, err = h.svc.GetPubByID(ctx, id)
		return er
	})

	uc := ctx.MustGet("user").(jwt.UserClaims)
	eg.Go(func() error {
		var er error
		inter, er = h.interSvc.Get(ctx, h.biz, id, uc.Uid)
		return er
	})

	// 等待结果
	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("查询文章失败，系统错误",
			logger.Int64("art_id", id),
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}

	go func() {
		// 1. 如果你想摆脱原本主链路的超时控制，你就创建一个新的
		// 2. 如果你不想，你就用 ctx
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := h.interSvc.IncrReadCnt(newCtx, h.biz, art.Id)
		if er != nil {
			h.l.Error("更新阅读数失败",
				logger.Int64("art_id", art.Id),
				logger.Error(err))
		}
	}()

	ctx.JSON(http.StatusOK, ginx.Result{
		Data: ArticleVo{
			Id:    art.Id,
			Title: art.Title,

			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			ReadCnt:    inter.ReadCnt,
			LikeCnt:    inter.LikeCnt,
			CollectCnt: inter.CollectCnt,
			Liked:      inter.Liked,
			Collected:  inter.Collected,

			Status: art.Status.ToUint8(),
			Ctime:  art.Ctime.Format(time.DateTime),
			Utime:  art.Utime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
		// true 是点赞，false 是不点赞
		Like bool `json:"like"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	var err error
	if req.Like {
		// 点赞
		err = h.interSvc.Like(ctx, h.biz, req.Id, uc.Uid)
	} else {
		err = h.interSvc.CancelLike(ctx, h.biz, req.Id, uc.Uid)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("点赞/取消点赞失败",
			logger.Error(err),
			logger.Int64("uid", uc.Uid),
			logger.Int64("art_id", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	err := h.interSvc.Collect(ctx, h.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5, Msg: "系统错误",
		})
		h.l.Error("收藏失败",
			logger.Error(err),
			logger.Int64("uid", uc.Uid),
			logger.Int64("aid", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}
