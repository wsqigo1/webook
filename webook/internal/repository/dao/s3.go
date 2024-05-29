package dao

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

type ArticleS3DAO struct {
	ArticleGORMDAO
	oss *s3.S3
}

func NewArticleS3DAO(db *gorm.DB, oss *s3.S3) *ArticleS3DAO {
	return &ArticleS3DAO{
		ArticleGORMDAO: ArticleGORMDAO{db: db},
		oss:            oss,
	}
}

func (a *ArticleS3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	id := art.Id
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		dao := NewArticleGORMDAO(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}

		art.Id = id
		now := time.Now().UnixMilli()
		// PublishedArticleV1 不具备 Content
		pubArt := PublishedArticleV2{
			Id:       art.Id,
			Title:    art.Title,
			AuthorId: art.AuthorId,
			Ctime:    now,
			Utime:    now,
			Status:   art.Status,
		}
		return tx.Clauses(clause.OnConflict{
			// 对MySQL不起效，但是可以兼容别的方言
			// INSERT xxx ON DUPLICATE KEY SET `title`=?
			// 别的方言：
			// sqlite INSERT XXX ON CONFLICT DO UPDATES WHERE
			//INSERT INTO xxx
			//ON CONFLICT (id)
			//DO UPDATE SET title = excluded.title, content = excluded.content, utime = excluded.utime;
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"title":  pubArt.Title,
				"utime":  pubArt.Utime,
				"status": pubArt.Status,
			}),
		}).Create(&pubArt).Error
	})
	if err != nil {
		return 0, err
	}
	// 最后同步到 OSS 上，但是只同步了 Content
	_, err = a.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("weibook"),
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text-plain;charset=utf-8"),
	})
	return id, err
}

func (a *ArticleS3DAO) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	now := time.Now().UnixMilli()
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id, uid).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("更新失败, ID不对或者作者不对")
		}
		return tx.Model(&PublishedArticleV2{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			}).Error
	})
	if err != nil {
		return err
	}
	const statusPrivate = 3
	if status == statusPrivate {
		_, err = a.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("weibook"),
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
	}
	return err
}

type PublishedArticleV2 struct {
	Id    int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	// 我要根据创作者ID来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
