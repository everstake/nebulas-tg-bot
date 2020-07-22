package bot

import (
	"fmt"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/models"
)

func (bot *Bot) setAddresses() error {
	addresses, err := bot.dao.GetAddresses(filters.Addresses{})
	if err != nil {
		return fmt.Errorf("dao.GetAddresses: %s", err.Error())
	}
	addressesMap := make(map[uint64]models.Address)
	for _, address := range addresses {
		addressesMap[address.ID] = address
	}
	users, err := bot.dao.GetUsers(filters.Users{})
	if err != nil {
		return fmt.Errorf("dao.GetUsers: %s", err.Error())
	}
	for _, user := range users {
		bot.users[user.ID] = user
	}
	userAddresses, err := bot.dao.GetUsersAddresses(filters.UsersAddresses{})
	if err != nil {
		return fmt.Errorf("dao.GetAddresses: %s", err.Error())
	}
	for _, ua := range userAddresses {
		address, _ := addressesMap[ua.AddressID]
		_, ok := bot.addresses[address.Address]
		if !ok {
			bot.addresses[address.Address] = make(map[uint64]struct{})
		}
		bot.addresses[address.Address][ua.UserID] = struct{}{}

		if ua.Type == models.AddressTypeValidator {
			_, ok = bot.validators[address.Address]
			if !ok {
				bot.validators[address.Address] = make(map[uint64]struct{})
			}
			bot.validators[address.Address][ua.UserID] = struct{}{}
		}
	}
	return nil
}

func (bot *Bot) updateUserSettings(user models.User) {
	bot.mu.Lock()
	bot.users[user.ID] = user
	bot.mu.Unlock()
}

func (bot *Bot) addValidatorAddress(user models.User, address models.Address) {
	bot.mu.Lock()
	_, ok := bot.validators[address.Address]
	if !ok {
		bot.validators[address.Address] = make(map[uint64]struct{})
	}
	bot.validators[address.Address][user.ID] = struct{}{}
	bot.mu.Unlock()
}

func (bot *Bot) addAccountAddress(user models.User, address models.Address) {
	bot.mu.Lock()
	_, ok := bot.addresses[address.Address]
	if !ok {
		bot.addresses[address.Address] = make(map[uint64]struct{})
	}
	bot.addresses[address.Address][user.ID] = struct{}{}
	bot.mu.Unlock()
}

func (bot *Bot) removeAddress(user models.User, address models.Address) {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	_, ok := bot.addresses[address.Address]
	if !ok {
		return
	}
	delete(bot.addresses[address.Address], user.ID)
}

func (bot *Bot) addressExist(address string) bool {
	bot.mu.RLock()
	_, ok := bot.addresses[address]
	bot.mu.RUnlock()
	return ok
}
