package service

import (
	"context"
	"errors"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, userId int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
	FindOrCreateByDDing(ctx context.Context, info domain.DDingInfo) (domain.User, error)
}

type userService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// DEBUG
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	return svc.repo.UpdateNonZeroFields(ctx, user)
}

func (svc *userService) FindById(ctx context.Context, userId int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, userId)
	return u, err
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 先找一下，我们认为，大部分用户是已经存在的用户
	// 这是一种优化写法
	// 大部分人会命中这个分支
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		// 有两种情况
		// err == nil, u 是可用的
		// err != nil，系统错误
		return u, err
	}
	// 用户没找到
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	// 有两种可能，一种是 err 恰好是唯一索引冲突（phone）
	// 一种是 err != nil，系统错误
	if err != nil && err != repository.ErrUserDuplicate {
		return domain.User{}, err
	}

	// 主从模式下，这里要从主库中读取，暂时我们不需要考虑
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	// 类似于手机号的过程，大部分人只是扫码登录，也就是数据在我们
	u, err := svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
	if err != repository.ErrUserNotFound {
		// 有两种情况
		// err == nil, u 是可用的
		// err != nil，系统错误
		return u, err
	}
	// 用户没找到
	// 这边就是意味着是一个新用户
	// JSON 格式的 wechatInfo
	// 直接使用包变量
	zap.L().Info("微信用户未注册，注册新用户",
		zap.Any("wechatInfo", wechatInfo))
	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: wechatInfo,
	})
	// 有两种可能，一种是 err 恰好是唯一索引冲突（phone）
	// 一种是 err != nil，系统错误
	if err != nil && err != repository.ErrUserDuplicate {
		return domain.User{}, err
	}

	// 主从模式下，这里要从主库中读取，暂时我们不需要考虑
	return svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
}

func (svc *userService) FindOrCreateByDDing(ctx context.Context, dDingInfo domain.DDingInfo) (domain.User, error) {
	u, err := svc.repo.FindByDDing(ctx, dDingInfo.OpenId)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	// 用户没找到
	zap.L().Info("新用户", zap.Any("dDingInfo", dDingInfo))
	err = svc.repo.Create(ctx, domain.User{
		DDingInfo: dDingInfo,
	})
	// 有两种可能，一种是 err 恰好是唯一索引冲突（phone）
	// 一种是 err != nil，系统错误
	if err != nil && err != repository.ErrUserDuplicate {
		return domain.User{}, err
	}

	// 主从模式下，这里要从主库中读取，暂时我们不需要考虑
	return svc.repo.FindByDDing(ctx, dDingInfo.OpenId)
}
