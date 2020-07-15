package dao

import (
	"fmt"
	"github.com/everstake/nebulas-tg-bot/config"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/dao/mysql"
	"github.com/everstake/nebulas-tg-bot/models"
)

type (
	DAO interface {
		Mysql
	}
	Mysql interface {
		GetUsers(filter filters.Users) (users []models.User, err error)
		CreateUser(user models.User) (models.User, error)
		UpdateUser(user models.User) error
	}

	daoImpl struct {
		Mysql
	}
)

func NewDAO(cfg config.Config) (DAO, error) {
	mysqlDB, err := mysql.NewDB(cfg.Mysql)
	if err != nil {
		return nil, fmt.Errorf("mysql.NewDB: %s", err.Error())
	}
	return daoImpl{
		Mysql: mysqlDB,
	}, nil
}
