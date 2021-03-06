package models

import (
	"github.com/shopspring/decimal"
	"time"
)

const UsersTable = "users"

type User struct {
	ID          uint64          `db:"usr_id"`
	TgID        int64           `db:"usr_tg_id"`
	Lang        string          `db:"usr_lang"`
	Username    string          `db:"usr_username"`
	Name        string          `db:"usr_name"`
	Mute        bool            `db:"usr_mute"`
	Step        string          `db:"usr_step"`
	MinThreshold decimal.Decimal `db:"usr_min_threshold"`
	MaxThreshold decimal.Decimal `db:"usr_max_threshold"`
	CreatedAt   time.Time       `db:"usr_created_at"`
}
