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

		GetAddresses(filter filters.Addresses) (addresses []models.Address, err error)
		CreateAddress(address models.Address) (models.Address, error)
		CreateUserAddress(userAddress models.UserAddress) error
		GetUsersAddresses(filter filters.UsersAddresses) (usersAddresses []models.UserAddress, err error)
		GetUsersAddressReports(filter filters.UsersAddresses) (items []models.UserAddressReport, err error)
		DeleteUserAddress(userID uint64, addressID uint64) error

		UpdateState(state models.State) error
		GetState(title string) (state models.State, err error)
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
