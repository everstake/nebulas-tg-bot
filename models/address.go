package models

import (
	"github.com/shopspring/decimal"
	"time"
)

const AddressesTable = "addresses"

type Address struct {
	ID        uint64    `db:"adr_id"`
	Address   string    `db:"adr_address"`
	CreatedAt time.Time `db:"adr_created_at"`
}

type AddressState struct {
	Address    string          `json:"address"`
	NAS        decimal.Decimal `json:"nas"`
	NAX        decimal.Decimal `json:"nax"`
	Alias      string          `json:"alias"`
	Type       string          `json:"type"`
	TotalVotes decimal.Decimal `json:"total_votes"`
}
