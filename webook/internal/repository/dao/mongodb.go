package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type ArticleMongoDBDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

//var _ ArticleDAO = &ArticleMongoDBDAO{}

func NewArticleMongoDBDAO(mdb *mongo.Database, node *snowflake.Node) *ArticleMongoDBDAO {
	return &ArticleMongoDBDAO{
		node:    node,
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
	}
}

func (m *ArticleMongoDBDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	art.Id = m.node.Generate().Int64()
	_, err := m.col.InsertOne(ctx, &art)
	return art.Id, err
}

func (m *ArticleMongoDBDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		// 这里你可以考虑直接使用整个 art，因为会忽略零值。
		// 参考 Sync 的中的写法
		// 但是我一般都喜欢显式指定要被更新的字段，确保可读性和可维护性
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		// 创作者不对，说明有人在瞎搞
		return errors.New("ID 不对或者创作者不对")
	}
	return nil
}

func (m *ArticleMongoDBDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	// liveCol
	// 是 INSERT or Update 语义
	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now
	filter := bson.D{bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId}}
	set := bson.D{bson.E{Key: "$set", Value: PublishedArticle(art)},
		bson.E{Key: "$setOnInsert", Value: bson.D{bson.E{Key: "ctime", Value: now}}}}

	_, err = m.liveCol.UpdateOne(ctx, filter, set,
		options.Update().SetUpsert(true))
	return id, err
}

func (m *ArticleMongoDBDAO) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{Key: "id", Value: id},
		bson.E{Key: "author_id", Value: uid}}
	sets := bson.D{bson.E{Key: "$set", Value: bson.D{
		bson.E{Key: "status", Value: status},
		bson.E{Key: "utime", Value: now},
	}}}
	res, err := m.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("更新失败, ID不对或者作者不对")
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, sets)
	return err
}
