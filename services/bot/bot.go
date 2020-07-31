package bot

import (
	"encoding/json"
	"fmt"
	"github.com/everstake/nebulas-tg-bot/config"
	"github.com/everstake/nebulas-tg-bot/dao"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/log"
	"github.com/everstake/nebulas-tg-bot/models"
	"github.com/everstake/nebulas-tg-bot/services/market"
	"github.com/everstake/nebulas-tg-bot/services/node"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"strings"
	"sync"
)

const StakingContract = "n214bLrE3nREcpRewHXF7qRDWCcaxRSiUdw"

type (
	Bot struct {
		cfg                  config.Config
		dao                  dao.DAO
		api                  *tgbotapi.BotAPI
		node                 NodeAPI
		market               marketAPI
		routes               map[string]Route
		dictionary           models.Dictionary
		cachedItems          map[uint64]map[string]interface{} // [userID][key]
		mu                   *sync.RWMutex
		addresses            map[string]map[uint64]struct{} // [address][userID]
		validators           map[string]map[uint64]struct{} // [address][userID]
		users                map[uint64]models.User
		nodes                map[string]node.ValidatorNode
		lastStabilityIndexes map[string]float64
	}
	marketAPI interface {
		GetNASPrice() decimal.Decimal
		GetNAXPrice() decimal.Decimal
		Run()
	}
	NodeAPI interface {
		GetAccountState(address string) (state node.AccountState, err error)
		GetBlock(height uint64) (block node.Block, err error)
		GetLatestIrreversibleBlock() (block node.Block, err error)
		GetNAXBalance(address string) (result decimal.Decimal, err error)
		GetNodesList() (list []node.ValidatorNode, err error)
		GetNodeVotesList(nodeID string) (list []node.Vote, err error)
		GetVotedNAX(address string) (amount decimal.Decimal, err error)
	}
)

func NewBot(d dao.DAO, cfg config.Config) *Bot {
	return &Bot{
		cfg:                  cfg,
		dao:                  d,
		cachedItems:          make(map[uint64]map[string]interface{}),
		market:               market.NewMarket(),
		node:                 node.NewAPI(cfg.Node),
		mu:                   &sync.RWMutex{},
		addresses:            make(map[string]map[uint64]struct{}),
		validators:           make(map[string]map[uint64]struct{}),
		users:                make(map[uint64]models.User),
		nodes:                make(map[string]node.ValidatorNode),
		lastStabilityIndexes: make(map[string]float64),
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

	err = bot.setAddresses()
	if err != nil {
		return fmt.Errorf("setAddresses: %s", err.Error())
	}

	err = bot.setNodes()
	if err != nil {
		return fmt.Errorf("setNodes: %s", err.Error())
	}

	go bot.market.Run()
	go bot.Parsing()

	bot.SetRoutes()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.api.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			err = bot.handleActions(update)
			if err != nil {
				log.Error("Bot: handleActions: %s", err.Error())
			}
		}
		if update.Message != nil {
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
	if update.Message.Text == "/start" {
		_ = bot.openRoute(RouteStart, user)
		return nil
	}
	route, ok := bot.routes[user.Step]
	if !ok {
		err = bot.routes[RouteStart].request(user)
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
	err = route.response(update, user)
	if err != nil {
		msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.oops", user.Lang))
		_, _ = bot.api.Send(msg)
		_ = bot.openRoute(RouteStart, user)
		return fmt.Errorf("route(response:%s): %s", user.Step, err.Error())
	}
	return nil
}

func (bot *Bot) handleActions(update tgbotapi.Update) error {
	users, err := bot.dao.GetUsers(filters.Users{TgIDs: []int64{update.CallbackQuery.Message.Chat.ID}})
	if err != nil {
		return fmt.Errorf("dao.GetUsers: %s", err.Error())
	}
	if len(users) == 0 {
		return nil
	}
	user := users[0]
	query := update.CallbackQuery.Data
	parts := strings.Split(query, "_")
	switch parts[0] {
	case "delete":
		if len(parts) == 1 {
			return nil
		}
		addresses, err := bot.dao.GetAddresses(filters.Addresses{Addresses: []string{parts[1]}})
		if err != nil {
			return fmt.Errorf("dao.GetAddresses: %s", err.Error())
		}
		if len(addresses) == 0 {
			return nil
		}
		err = bot.dao.DeleteUserAddress(user.ID, addresses[0].ID)
		if err != nil {
			return fmt.Errorf("dao.GetAddresses: %s", err.Error())
		}
		_, err = bot.api.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    user.TgID,
			MessageID: update.CallbackQuery.Message.MessageID,
		})
		if err != nil {
			return fmt.Errorf("api.DeleteMessage: %s", err.Error())
		}
		bot.removeAddress(user, addresses[0])
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
			TgID:        tgID,
			Name:        update.Message.Chat.FirstName + " " + update.Message.Chat.LastName,
			Username:    update.Message.Chat.UserName,
			Lang:        "en",
			MaxThreshold: decimal.NewFromFloat(99999999999),
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
