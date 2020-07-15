package models

import "time"

const AddressesTable = "addresses"

type Address struct {
	ID        uint64    `db:"adr_id"`
	Address   string    `db:"adr_address"`
	CreatedAt time.Time `db:"adr_created_at"`
}
