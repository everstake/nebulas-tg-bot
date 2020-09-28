package bot

import (
	"fmt"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/log"
	"github.com/everstake/nebulas-tg-bot/models"
	"github.com/everstake/nebulas-tg-bot/services/node"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/shopspring/decimal"
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
	nasPrice := bot.market.GetNASPrice()
	naxPrice := bot.market.GetNAXPrice()
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
		var text string
		switch state.Type {
		case models.AddressTypeAccount:
			text = fmt.Sprintf(
				bot.dictionary.Get("t.address_subscription", user.Lang),
				state.Alias,
				state.Address,
				state.NAS.Truncate(4).String(),
				state.NAS.Mul(nasPrice).Truncate(4).String(),
				state.NAX.Truncate(4).String(),
				state.NAX.Mul(naxPrice).Truncate(6).String(),
				state.Type,
				state.VotedAmount.Truncate(4).String(),
			)
		case models.AddressTypeValidator:
			text = fmt.Sprintf(
				bot.dictionary.Get("t.validator_subscription", user.Lang),
				state.Alias,
				state.Address,
				state.NAS.Truncate(4).String(),
				state.NAS.Mul(nasPrice).Truncate(4).String(),
				state.NAX.Truncate(4).String(),
				state.NAX.Mul(naxPrice).Truncate(6).String(),
				state.Type,
				state.TotalVotes,
			)
		default:
			continue
		}

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
			naxBalance, err := bot.node.GetNAXBalance(address.Address)
			if err != nil {
				errChan <- fmt.Errorf("node.GetNAXBalance: %s", err.Error())
				return
			}
			totalVotes := decimal.Zero
			votedAmount := decimal.Zero
			if address.Type == models.AddressTypeValidator {
				var nodeID string
				bot.mu.RLock()
				for _, n := range bot.nodes {
					if n.Accounts.StakingAccount == address.Address ||
						n.Accounts.Registrant == address.Address ||
						n.Accounts.GovManager == address.Address ||
						n.Accounts.ConsensusManager == address.Address {

						nodeID = n.ID
						break
					}
				}
				bot.mu.RUnlock()
				if nodeID != "" {
					list, err := bot.node.GetNodeVotesList(nodeID)
					if err != nil {
						log.Error("getSubscriptions: node.GetNodeVotesList: %s", err.Error())
					} else {
						for _, vote := range list {
							totalVotes = totalVotes.Add(vote.Value)
						}
						totalVotes = totalVotes.Div(node.PrecisionDivNAX)
					}
				}
			} else {
				votedAmount, err = bot.node.GetVotedNAX(address.Address)
				if err != nil {
					log.Error("getSubscriptions: node.GetVotedNAX: %s", err.Error())
				} else {
					votedAmount = votedAmount.Div(node.PrecisionDivNAX)
				}
			}
			stateCh <- models.AddressState{
				Address:     address.Address,
				NAS:         as.Result.Balance.Div(node.PrecisionDivNAS),
				NAX:         naxBalance.Div(node.PrecisionDivNAX),
				Alias:       address.Alias,
				Type:        address.Type,
				TotalVotes:  totalVotes,
				VotedAmount: votedAmount,
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

func getUniqStrings(items []string) []string {
	var nItems []string
	for _, item := range items {
		found := false
		for _, s := range nItems {
			if s == item {
				found = true
				break
			}
		}
		if !found {
			nItems = append(nItems, item)
		}
	}
	return nItems

}

func (bot *Bot) sendMsg(msg tgbotapi.Chattable) error {
	_, err := bot.api.Send(msg)
	if err != nil {
		if err.Error() == BlockedByUserErr {
			return nil
		}
	}
	return err
}
