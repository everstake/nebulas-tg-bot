package mysql

import (
	"github.com/Masterminds/squirrel"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/models"
)

func (m DB) GetUsers(filter filters.Users) (users []models.User, err error) {
	q := squirrel.Select("*").From(models.UsersTable)
	if len(filter.IDs) != 0 {
		q = q.Where(squirrel.Eq{"usr_id": filter.IDs})
	}
	if len(filter.TgIDs) != 0 {
		q = q.Where(squirrel.Eq{"usr_tg_id": filter.TgIDs})
	}
	err = m.find(&users, q)
	return users, err
}

func (m DB) CreateUser(user models.User) (models.User, error) {
	q := squirrel.Insert(models.UsersTable).SetMap(map[string]interface{}{
		"usr_tg_id":        user.TgID,
		"usr_name":         user.Name,
		"usr_lang":         user.Lang,
		"usr_username":     user.Username,
		"usr_mute":         user.Mute,
		"usr_step":         user.Step,
		"usr_min_threshold": user.MinThreshold,
		"usr_max_threshold": user.MaxThreshold,
	})
	var err error
	user.ID, err = m.insert(q)
	return user, err
}

func (m DB) UpdateUser(user models.User) error {
	q := squirrel.Update(models.UsersTable).SetMap(map[string]interface{}{
		"usr_lang":         user.Lang,
		"usr_mute":         user.Mute,
		"usr_step":         user.Step,
		"usr_min_threshold": user.MinThreshold,
		"usr_max_threshold": user.MaxThreshold,
	}).Where(squirrel.Eq{"usr_id": user.ID})
	return m.update(q)
}
