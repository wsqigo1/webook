//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/wsqigo/basic-go/wire/repository"
	"github.com/wsqigo/basic-go/wire/repository/dao"
)

func InitUserRepository() *repository.UserRepository {
	wire.Build(repository.NewUserRepository, dao.NewUserDAO, InitDB)
	return &repository.UserRepository{}
}
