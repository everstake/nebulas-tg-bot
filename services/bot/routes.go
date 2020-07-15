package bot

import (
	"fmt"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

const (
	RouteStart        = "start"
	RouteChooseLang   = "choose_lang"
	RouteSettings     = "settings"
	RouteAddAddress   = "add_address"
	RouteAddressAlias = "address_alias"
)

type Route struct {
	request  func(user models.User) error
	response func(update tgbotapi.Update, user models.User) error
}

func (bot *Bot) SetRoutes() {
	bot.routes = map[string]Route{
		RouteChooseLang: {
			request: func(user models.User) error {
				var numericKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.lang_en", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.lang_cn", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.choose_lang", user.Lang))
				msg.ReplyMarkup = numericKeyboard
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				lang := "en"
				switch update.Message.Text {
				case bot.dictionary.Get("b.lang_en", user.Lang):
					lang = "en"
				case bot.dictionary.Get("b.lang_cn", user.Lang):
					lang = "cn"
				default:
					msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.wrong_lang", user.Lang))
					_, err := bot.api.Send(msg)
					if err != nil {
						return fmt.Errorf("api.Send: %s", err.Error())
					}
				}
				user.Lang = lang
				err := bot.openRoute(RouteStart, user)
				if err != nil {
					return fmt.Errorf("openRoute: %s", err.Error())
				}
				return nil
			},
		},
		RouteStart: {
			request: func(user models.User) error {
				var numericKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.add_address", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.show_addresses", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.settings", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.menu", user.Lang))
				msg.ReplyMarkup = numericKeyboard
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				switch update.Message.Text {
				case bot.dictionary.Get("b.settings", user.Lang):
					err := bot.openRoute(RouteSettings, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
				case bot.dictionary.Get("b.add_address", user.Lang):
					err := bot.openRoute(RouteAddAddress, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
				default:
					err := bot.openRoute(RouteStart, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
				}
				return nil
			},
		},
		RouteSettings: {
			request: func(user models.User) error {
				var numericKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.change_lang", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.return_back", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.choose_option", user.Lang))
				msg.ReplyMarkup = numericKeyboard
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				switch update.Message.Text {
				case bot.dictionary.Get("b.change_lang", user.Lang):
					err := bot.openRoute("choose_lang", user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
				case bot.dictionary.Get("b.return_back", user.Lang):
					err := bot.openRoute(RouteStart, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
				default:
					msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.wrong_option", user.Lang))
					_, err := bot.api.Send(msg)
					if err != nil {
						return fmt.Errorf("api.Send: %s", err.Error())
					}
				}
				return nil
			},
		},
		RouteAddAddress: {
			request: func(user models.User) error {
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.paste_your_address", user.Lang))
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				text := update.Message.Text
				text = strings.TrimSpace(text)
				if len(text) != 35 || text[0] != 'n' {
					msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.wrong_address", user.Lang))
					_, err := bot.api.Send(msg)
					if err != nil {
						return fmt.Errorf("api.Send: %s", err.Error())
					}
					return nil
				}

				addresses, err := bot.dao.GetAddresses(filters.Addresses{Addresses: []string{text}})
				if err != nil {
					return fmt.Errorf("dao.GetAddresses: %s", err.Error())
				}
				if len(addresses) != 0 {
					usersAddresses, err := bot.dao.GetUsersAddresses(filters.UsersAddresses{
						AddressesID: []uint64{addresses[0].ID},
						UserID:      []uint64{user.ID}},
					)
					if err != nil {
						return fmt.Errorf("dao.GetUsersAddresses: %s", err.Error())
					}
					if len(usersAddresses) != 0 {
						msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.address_already_added", user.Lang))
						_, err = bot.api.Send(msg)
						if err != nil {
							return fmt.Errorf("api.Send: %s", err.Error())
						}
						err = bot.openRoute(RouteStart, user)
						if err != nil {
							return fmt.Errorf("openRoute: %s", err.Error())
						}
						return nil
					}
				}

				bot.SetCachedItem(user.ID, "address", text)
				err = bot.openRoute(RouteAddressAlias, user)
				if err != nil {
					return fmt.Errorf("openRoute: %s", err.Error())
				}
				return nil
			},
		},
		RouteAddressAlias: {
			request: func(user models.User) error {
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.enter_address_alias", user.Lang))
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				item, ok := bot.GetCachedItem(user.ID, "address")
				if !ok {
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
				address := item.(string)
				addresses, err := bot.dao.GetAddresses(filters.Addresses{Addresses: []string{address}})
				if err != nil {
					return fmt.Errorf("dao.GetAddresses: %s", err.Error())
				}
				var addressModel models.Address
				if len(addresses) == 0 {
					addressModel, err = bot.dao.CreateAddress(models.Address{
						Address: address,
					})
					if err != nil {
						return fmt.Errorf("dao.CreateAddress: %s", err.Error())
					}
				} else {
					addressModel = addresses[0]
				}
				alias := update.Message.Text
				if len(alias) > 100 {
					alias = alias[:100]
				}
				err = bot.dao.CreateUserAddress(models.UserAddress{
					UserID:    user.ID,
					AddressID: addressModel.ID,
					Alias:     alias,
				})
				if err != nil {
					return fmt.Errorf("dao.CreateUserAddress: %s", err.Error())
				}
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.address_added", user.Lang))
				_, err = bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				err = bot.openRoute(RouteStart, user)
				if err != nil {
					return fmt.Errorf("openRoute: %s", err.Error())
				}
				return nil
			},
		},
	}
}

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
