package bot

import (
	"fmt"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/models"
	"github.com/everstake/nebulas-tg-bot/services/node"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (bot *Bot) GetCachedItem(userID uint64, key string) (item interface{}, found bool) {
	mp, ok := bot.cachedItems[userID]
	if !ok {
		bot.cachedItems[userID] = make(map[string]interface{})
	}
	item, ok = mp[key]
	return item, ok
}

func (bot *Bot) SetCachedItem(userID uint64, key string, item interface{}) {
	_, ok := bot.cachedItems[userID]
	if !ok {
		bot.cachedItems[userID] = make(map[string]interface{})
	}
	bot.cachedItems[userID][key] = item
}

func (bot *Bot) ClearCachedItems(userID uint64) {
	delete(bot.cachedItems, userID)
}

func (bot *Bot) openRoute(key string, user models.User) error {
	user.Step = key
	err := bot.dao.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("dao.UpdateUser: %s", err.Error())
	}
	route, ok := bot.routes[key]
	if !ok {
		return fmt.Errorf("not found route %s", key)
	}
	err = route.request(user)
	if err != nil {
		return fmt.Errorf("actions(%s): %s", key, err.Error())
	}
	return nil
}

func (bot *Bot) oops(user models.User) error {
	msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.oops", user.Lang))
	_, err := bot.api.Send(msg)
	if err != nil {
		return fmt.Errorf("api.Send: %s", err.Error())
	}
	err = bot.openRoute(RouteStart, user)
	if err != nil {
		return fmt.Errorf("openRoute: %s", err.Error())
	}
	return nil
}

func (bot *Bot) showSubscriptions(user models.User) (err error) {
	price := bot.market.GetPrice()
	states, err := bot.getSubscriptions(user)
	if err != nil {
		return fmt.Errorf("getSubscriptions: %s", err.Error())
	}
	if len(states) == 0 {
		msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.not_have_addresses", user.Lang))
		_, err := bot.api.Send(msg)
		if err != nil {
			return fmt.Errorf("api.Send: %s", err.Error())
		}
		return nil
	}
	for _, state := range states {
		text := fmt.Sprintf(
			"Alias: %s\nAddress: %s\nNAS: %s (%s$)\nType: %s",
			state.Alias,
			state.Address,
			state.NAS.Truncate(4).String(),
			state.NAS.Mul(price).Truncate(0).String(),
			state.Type,
		)
		url := fmt.Sprintf("https://explorer.nebulas.io/#/address/%s", state.Address)
		action := fmt.Sprintf("delete_%s", state.Address)
		var keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(bot.dictionary.Get("b.link", user.Lang), url),
				tgbotapi.NewInlineKeyboardButtonData(bot.dictionary.Get("b.delete", user.Lang), action),
			),
		)
		msg := tgbotapi.NewMessage(user.TgID, text)
		msg.ReplyMarkup = keyboard
		_, err := bot.api.Send(msg)
		if err != nil {
			return fmt.Errorf("api.Send: %s", err.Error())
		}
	}
	return nil
}

func (bot *Bot) getSubscriptions(user models.User) (states []models.AddressState, err error) {
	addresses, err := bot.dao.GetUsersAddressReports(filters.UsersAddresses{
		UserID: []uint64{user.ID},
		Limit:  20,
	})
	if err != nil {
		return nil, fmt.Errorf("dao.GetUsersAddressReports: %s", err.Error())
	}
	if len(addresses) == 0 {
		return nil, nil
	}
	errChan := make(chan error)
	stateCh := make(chan models.AddressState)
	defer close(errChan)
	defer close(stateCh)

	for i := range addresses {
		go func(i int) {
			address := addresses[i]
			as, err := bot.node.GetAccountState(address.Address)
			if err != nil {
				errChan <- fmt.Errorf("node.GetAccountState: %s", err.Error())
				return
			}
			stateCh <- models.AddressState{
				Address: address.Address,
				NAS:     as.Result.Balance.Div(node.PrecisionDiv),
				Alias:   address.Alias,
				Type:    address.Type,
			}
		}(i)
	}

	for {
		exit := false
		select {
		case err = <-errChan:
			return nil, err
		case s := <-stateCh:
			states = append(states, s)
			if len(states) == len(addresses) {
				exit = true
			}
		}
		if exit {
			break
		}
	}

	return states, nil
}
