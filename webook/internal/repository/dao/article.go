package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
}

type ArticleGORMDAO struct {
	db *gorm.DB
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{
		db: db,
	}
}

func (a *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
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
		pubArt := PublishedArticle(art)
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			// 对MySQL不起效，但是可以兼容别的方言
			// INSERT xxx ON DUPLICATE KEY SET `title`=?
			// 别的方言：
			// sqlite INSERT XXX ON CONFLICT DO UPDATES WHERE
			//INSERT INTO xxx
			//ON CONFLICT (id)
			//DO UPDATE SET title = excluded.title, content = excluded.content, utime = excluded.utime;
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"utime":   pubArt.Utime,
			}),
		}).Create(&pubArt).Error
		return err
	})

	return id, err
}

func (a *ArticleGORMDAO) SyncV1(ctx context.Context, art Article) (int64, error) {
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, nil
	}

	// 防止后面业务panic
	// 直接 defer Rollback
	// 如果我们后续 Commit 了，这里会得到一个错误，但是没关系
	defer tx.Rollback()

	var (
		id  = art.Id
		err error
	)

	dao := NewArticleGORMDAO(tx)
	if id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	art.Id = id
	now := time.Now().UnixMilli()
	pubArt := PublishedArticle(art)
	pubArt.Ctime = now
	pubArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		// 对MySQL不起效，但是可以兼容别的方言
		// INSERT xxx ON DUPLICATE KEY SET `title`=?
		// 别的方言：
		// sqlite INSERT XXX ON CONFLICT DO UPDATES WHERE
		//INSERT INTO xxx
		//ON CONFLICT (id)
		//DO UPDATE SET title = excluded.title, content = excluded.content, utime = excluded.utime;
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"title":   pubArt.Title,
			"content": pubArt.Content,
			"utime":   pubArt.Utime,
		}),
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}

func (a *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&Article{}).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("更新失败, ID不对或者作者不对")
	}
	return nil
}

func (a *ArticleGORMDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 标题的长度
	// 正常都不会超过这个长度
	Title   string `gorm:"type=varchar(4096)"`
	Content string `gorm:"type=BLOB"`
	// 作者
	// 我要根据创作者ID来查询
	AuthorId int64 `gorm:"index"`
	Status   int64
	Ctime    int64
	// 更新时间
	Utime int64
}
