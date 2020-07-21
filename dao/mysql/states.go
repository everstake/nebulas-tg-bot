package mysql

import (
	"github.com/Masterminds/squirrel"
	"github.com/everstake/nebulas-tg-bot/models"
)

func (m DB) UpdateState(state models.State) error {
	q := squirrel.Update(models.StatesTable).
		Where(squirrel.Eq{"stt_title": state.Title}).
		SetMap(map[string]interface{}{
			"stt_value": state.Value,
		})

	err := m.update(q)
	return err
}

func (m DB) GetState(title string) (state models.State, err error) {
	q := squirrel.Select("*").From(models.StatesTable).Where(squirrel.Eq{"stt_title": title})
	err = m.first(&state, q)
	return state, err
}
