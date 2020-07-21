package models

import "time"

const UserAddressesTable = "users_addresses"

type UserAddress struct {
	UserID    uint64 `db:"usr_id"`
	AddressID uint64 `db:"adr_id"`
	Alias     string `db:"usa_alias"`
	Type      string `db:"usa_type"`
}

type UserAddressReport struct {
	ID        uint64    `db:"id"`
	Address   string    `db:"address"`
	Alias     string    `db:"alias"`
	Type      string    `db:"type"`
	CreatedAt time.Time `db:"created_at"`
}
