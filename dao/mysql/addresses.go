package mysql

import (
	"github.com/Masterminds/squirrel"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/models"
)

func (m DB) GetAddresses(filter filters.Addresses) (addresses []models.Address, err error) {
	q := squirrel.Select("*").From(models.AddressesTable)
	if len(filter.Addresses) != 0 {
		q = q.Where(squirrel.Eq{"adr_address": filter.Addresses})
	}
	err = m.find(&addresses, q)
	return addresses, err
}

func (m DB) CreateAddress(address models.Address) (models.Address, error) {
	q := squirrel.Insert(models.AddressesTable).SetMap(map[string]interface{}{
		"adr_address": address.Address,
	})
	var err error
	address.ID, err = m.insert(q)
	return address, err
}

func (m DB) CreateUserAddress(userAddress models.UserAddress) error {
	q := squirrel.Insert(models.UserAddressesTable).SetMap(map[string]interface{}{
		"usr_id":    userAddress.UserID,
		"adr_id":    userAddress.AddressID,
		"usa_alias": userAddress.Alias,
	})
	_, err := m.insert(q)
	return err
}

func (m DB) GetUsersAddresses(filter filters.UsersAddresses) (usersAddresses []models.UserAddress, err error) {
	q := squirrel.Select("*").From(models.UserAddressesTable)
	if len(filter.UserID) != 0 {
		q = q.Where(squirrel.Eq{"usr_id": filter.UserID})
	}
	if len(filter.AddressesID) != 0 {
		q = q.Where(squirrel.Eq{"adr_id": filter.AddressesID})
	}
	err = m.find(&usersAddresses, q)
	return usersAddresses, err
}
