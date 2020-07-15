package models

const UserAddressesTable = "users_addresses"

type UserAddress struct {
	UserID    uint64 `db:"usr_id"`
	AddressID uint64 `db:"adr_id"`
	Alias     string `db:"usa_alias"`
}
