package bot

import (
	"encoding/json"
	"fmt"
	"github.com/everstake/nebulas-tg-bot/config"
	"github.com/everstake/nebulas-tg-bot/dao"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/log"
	"github.com/everstake/nebulas-tg-bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
)

type Bot struct {
	cfg        config.Config
	dao        dao.DAO
	api        *tgbotapi.BotAPI
	routes     map[string]Route
	dictionary models.Dictionary
}

func NewBot(d dao.DAO, cfg config.Config) *Bot {
	return &Bot{
		cfg: cfg,
		dao: d,
	}
}

func (bot *Bot) Run() (err error) {
	bot.api, err = tgbotapi.NewBotAPI(bot.cfg.TelegramToken)
	if err != nil {
		return fmt.Errorf("tgbotapi.NewBotAPI: %s", err.Error())
	}

	data, err := ioutil.ReadFile("./dictionary.json")
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile: %s", err.Error())
	}
	err = json.Unmarshal(data, &bot.dictionary)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %s", err.Error())
	}

	bot.SetRoutes()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.api.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			fmt.Println(update.CallbackQuery.Data)
			fmt.Println(update.CallbackQuery.Message.Text)
		}
		if update.Message != nil { // ignore any non-Message Updates
			err = bot.handleUpdate(update)
			if err != nil {
				log.Error("Bot: handleUpdate: %s", err.Error())
			}
		}
	}
	return fmt.Errorf("updates has been broken")
}

func (bot *Bot) Stop() (err error) {
	return nil
}

func (bot *Bot) Title() string {
	return "Telegram Bot"
}

func (bot *Bot) handleUpdate(update tgbotapi.Update) error {
	user, err := bot.findOrCreateUser(update)
	if err != nil {
		return fmt.Errorf("findOrCreateUser: %s", err.Error())
	}

	if update.Message != nil {
		route, ok := bot.routes[user.Step]
		if !ok {
			err = bot.routes[RouteStart].Request(user)
			if err != nil {
				return fmt.Errorf("route(request:%s): %s", RouteStart, err.Error())
			}
			user.Step = RouteStart
			err = bot.dao.UpdateUser(user)
			if err != nil {
				return fmt.Errorf("dao.UpdateUser: %s", err.Error())
			}
			return nil
		}
		err = route.Response(update, user)
		if err != nil {
			return fmt.Errorf("route(response:%s): %s", user.Step, err.Error())
		}
		return nil
	}

	return nil
}

func (bot *Bot) findOrCreateUser(update tgbotapi.Update) (user models.User, err error) {
	tgID := update.Message.Chat.ID
	users, err := bot.dao.GetUsers(filters.Users{TgIDs: []int64{tgID}})
	if err != nil {
		return user, fmt.Errorf("dao.GetUsers: %s", err.Error())
	}
	if len(users) == 0 {
		user, err = bot.dao.CreateUser(models.User{
			TgID:     tgID,
			Name:     update.Message.Chat.FirstName + " " + update.Message.Chat.LastName,
			Username: update.Message.Chat.UserName,
		})
		if err != nil {
			return user, fmt.Errorf("dao.CreateUser: %s", err.Error())
		}
	} else {
		user = users[0]
	}
	return user, nil
}

func (bot *Bot) updateUser(update tgbotapi.Update) (user models.User, err error) {
	tgID := update.Message.Chat.ID
	users, err := bot.dao.GetUsers(filters.Users{TgIDs: []int64{tgID}})
	if err != nil {
		return user, fmt.Errorf("dao.GetUsers: %s", err.Error())
	}
	if len(users) == 0 {
		user, err = bot.dao.CreateUser(models.User{
			TgID:     tgID,
			Name:     update.Message.Chat.FirstName + " " + update.Message.Chat.LastName,
			Username: update.Message.Chat.UserName,
		})
		if err != nil {
			return user, fmt.Errorf("dao.CreateUser: %s", err.Error())
		}
	} else {
		user = users[0]
	}
	return user, nil
}
