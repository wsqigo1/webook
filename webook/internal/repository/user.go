package repository

import (
	"context"
	"database/sql"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/repository/cache"
	"github.com/wsqigo/basic-go/webook/internal/repository/dao"
	"log"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrUserNotFount
)

type UserRepository struct {
	cache *cache.UserCache
	dao   *dao.UserDAO
}

//type DBConfig struct {
//	DSN string
//}

//type CacheConfig struct {
//	Addr string
//}

// NewUserRepositoryV2 强耦合到了 JSON
//func NewUserRepositoryV2(cfgBytes string) *CachedUserRepository {
//	var cfg DBConfig
//	err := json.Unmarshal([]byte(cfgBytes), &cfg)
//}

// NewUserRepositoryV1 强耦合（跨层的），严重缺乏扩展性
//func NewUserRepositoryV1(dbCfg DBConfig, cCfg CacheConfig) (*CachedUserRepository, error) {
//	db, err := gorm.Open(mysql.Open(dbCfg.DSN))
//	if err != nil {
//		return nil, err
//	}
//	ud := dao.NewUserDAO(db)
//	uc := cache.NewUserCache(redis.NewClient(&redis.Options{
//		Addr: cCfg.Addr,
//	}))
//	return &CachedUserRepository{
//		dao:   ud,
//		cache: uc,
//	}, nil
//}

func NewUserRepository(d *dao.UserDAO, c *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   d,
		cache: c,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.toEntity(u))
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return repo.toDomain(u), nil
}

func (repo *UserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: time.UnixMilli(u.Birthday),
		AboutMe:  u.AboutMe,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}

func (repo *UserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Birthday: u.Birthday.UnixMilli(),
		AboutMe:  u.AboutMe,
		Nickname: u.Nickname,
	}
}

func (repo *UserRepository) UpdateNonZeroFields(ctx context.Context, user domain.User) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(user))
}

func (repo *UserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, uid)
	// 只要 err 为 nil，就返回
	if err == nil {
		return du, err
	}

	// err 不为 nil，就要查询数据库
	// err 有两种可能
	// 1. key 不存在，说明 redis 是正常的
	// 2. 访问 redis 有问题。可能是网络有问题，也可能是 redis 本身就崩溃了
	u, err := repo.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	du = repo.toDomain(u)
	//go func() {
	//	err = repo.cache.Set(ctx, du)
	//	if err != nil {
	//		log.Println(err)
	//	}
	//}()
	err = repo.cache.Set(ctx, du)
	if err != nil {
		// 网络崩了，也可能是 redis 崩了
		log.Println(err)
	}

	return du, nil
}

func (repo *UserRepository) FindByIdV2(ctx context.Context, uid int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, uid)
	switch err {
	case nil:
		return du, nil
	case cache.ErrKeyNotExist:
		u, err := repo.dao.FindById(ctx, uid)
		if err != nil {
			return domain.User{}, err
		}
		du = repo.toDomain(u)
		err = repo.cache.Set(ctx, du)
		if err != nil {
			// 网络崩了，也可能是 redis 崩了
			log.Println(err)
		}

		return du, nil
	default:
		// 接近降级的写法
		return domain.User{}, err
	}
}

func (repo *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	du := repo.toDomain(u)
	return du, nil
}
