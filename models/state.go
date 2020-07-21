package models

const StatesTable = "states"

const (
	StateCurrentHeight = "current_height"
)

type State struct {
	Title string `db:"stt_title"`
	Value string `db:"stt_value"`
}
