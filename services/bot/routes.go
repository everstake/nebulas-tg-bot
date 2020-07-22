package bot

import (
	"fmt"
	"github.com/everstake/nebulas-tg-bot/dao/filters"
	"github.com/everstake/nebulas-tg-bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/shopspring/decimal"
	"strings"
)

const (
	RouteStart          = "start"
	RouteChooseLang     = "choose_lang"
	RouteSettings       = "settings"
	RouteTypeAddress    = "type_address"
	RoutePasteAddress   = "paste_address"
	RouteAddressAlias   = "address_alias"
	RouteChangeTreshold = "change_treshold"
)

type Route struct {
	request  func(user models.User) error
	response func(update tgbotapi.Update, user models.User) error
}

func (bot *Bot) SetRoutes() {
	bot.routes = map[string]Route{
		RouteChooseLang: {
			request: func(user models.User) error {
				var keyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.lang_en", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.lang_cn", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.choose_lang", user.Lang))
				msg.ReplyMarkup = keyboard
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
				bot.updateUserSettings(user)
				err := bot.openRoute(RouteStart, user)
				if err != nil {
					return fmt.Errorf("openRoute: %s", err.Error())
				}
				return nil
			},
		},
		RouteStart: {
			request: func(user models.User) error {
				var keyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.add_subscription", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.show_subscriptions", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.settings", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.menu", user.Lang))
				msg.ReplyMarkup = keyboard
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
				case bot.dictionary.Get("b.add_subscription", user.Lang):
					err := bot.openRoute(RouteTypeAddress, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
				case bot.dictionary.Get("b.show_subscriptions", user.Lang):
					err := bot.showSubscriptions(user)
					if err != nil {
						return fmt.Errorf("showAddresses: %s", err.Error())
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
				muteButton := tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.mute", user.Lang))
				if user.Mute {
					muteButton = tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.unmute", user.Lang))
				}
				var keyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						muteButton,
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.change_lang", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.change_treshold", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.return_back", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.choose_option", user.Lang))
				msg.ReplyMarkup = keyboard
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
				case bot.dictionary.Get("b.mute", user.Lang), bot.dictionary.Get("b.unmute", user.Lang):
					user.Mute = true
					if update.Message.Text == bot.dictionary.Get("b.unmute", user.Lang) {
						user.Mute = false
					}
					err := bot.dao.UpdateUser(user)
					if err != nil {
						return fmt.Errorf("dao.UpdateUser: %s", err.Error())
					}
					bot.updateUserSettings(user)
					btnText := "t.muted"
					if !user.Mute {
						btnText = "t.unmuted"
					}
					msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get(btnText, user.Lang))
					_, err = bot.api.Send(msg)
					if err != nil {
						return fmt.Errorf("api.Send: %s", err.Error())
					}
					err = bot.openRoute(RouteSettings, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
				case bot.dictionary.Get("b.change_treshold", user.Lang):
					err := bot.openRoute(RouteChangeTreshold, user)
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
		RouteTypeAddress: {
			request: func(user models.User) error {
				var keyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.account_address", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.validator_address", user.Lang)),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.cancel", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.choose_address_type", user.Lang))
				msg.ReplyMarkup = keyboard
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				if update.Message.Text == bot.dictionary.Get("b.cancel", user.Lang) {
					err := bot.openRoute(RouteStart, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
					return nil
				}
				var err error
				switch update.Message.Text {
				case bot.dictionary.Get("b.account_address", user.Lang):
					bot.SetCachedItem(user.ID, "type_address", "account")
				case bot.dictionary.Get("b.validator_address", user.Lang):
					bot.SetCachedItem(user.ID, "type_address", "validator")
				default:
					msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.wrong_option", user.Lang))
					_, err := bot.api.Send(msg)
					if err != nil {
						return fmt.Errorf("api.Send: %s", err.Error())
					}
					return nil
				}
				err = bot.openRoute(RoutePasteAddress, user)
				if err != nil {
					return fmt.Errorf("openRoute: %s", err.Error())
				}
				return nil
			},
		},
		RoutePasteAddress: {
			request: func(user models.User) error {
				var keyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.cancel", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.paste_your_address", user.Lang))
				msg.ReplyMarkup = keyboard
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				text := update.Message.Text
				if text == bot.dictionary.Get("b.cancel", user.Lang) {
					err := bot.openRoute(RouteStart, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
					return nil
				}
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
				var keyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.cancel", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.enter_address_alias", user.Lang))
				msg.ReplyMarkup = keyboard
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				if update.Message.Text == bot.dictionary.Get("b.cancel", user.Lang) {
					err := bot.openRoute(RouteStart, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
					return nil
				}
				item, ok := bot.GetCachedItem(user.ID, "address")
				if !ok {
					return bot.oops(user)
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
				itemTypeAddress, ok := bot.GetCachedItem(user.ID, "type_address")
				if !ok {
					return bot.oops(user)
				}
				err = bot.dao.CreateUserAddress(models.UserAddress{
					UserID:    user.ID,
					AddressID: addressModel.ID,
					Alias:     alias,
					Type:      itemTypeAddress.(string),
				})
				if err != nil {
					return fmt.Errorf("dao.CreateUserAddress: %s", err.Error())
				}

				switch itemTypeAddress.(string) {
				case models.AddressTypeValidator:
					bot.addValidatorAddress(user, addressModel)
				case models.AddressTypeAccount:
					bot.addValidatorAddress(user, addressModel)
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
		RouteChangeTreshold: {
			request: func(user models.User) error {
				var keyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(bot.dictionary.Get("b.return_back", user.Lang)),
					),
				)
				msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.paste_treshold", user.Lang))
				msg.ReplyMarkup = keyboard
				_, err := bot.api.Send(msg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				return nil
			},
			response: func(update tgbotapi.Update, user models.User) error {
				msg := update.Message.Text
				if msg == bot.dictionary.Get("b.return_back", user.Lang) {
					err := bot.openRoute(RouteSettings, user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
					return nil
				}
				msg = strings.TrimSpace(msg)
				parts := strings.Split(msg, " ")
				if len(parts) != 2 {
					msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.invalid_treshold", user.Lang))
					_, err := bot.api.Send(msg)
					if err != nil {
						return fmt.Errorf("api.Send: %s", err.Error())
					}
					return nil
				}
				min, minErr := decimal.NewFromString(parts[0])
				max, maxErr := decimal.NewFromString(parts[1])
				if minErr != nil || maxErr != nil {
					msg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.invalid_treshold", user.Lang))
					_, err := bot.api.Send(msg)
					if err != nil {
						return fmt.Errorf("api.Send: %s", err.Error())
					}
					return nil
				}
				user.MinTreshold = min
				user.MaxTreshold = max
				err := bot.dao.UpdateUser(user)
				if err != nil {
					return fmt.Errorf("dao.UpdateUser: %s", err.Error())
				}
				tgMsg := tgbotapi.NewMessage(user.TgID, bot.dictionary.Get("t.successful_updated", user.Lang))
				_, err = bot.api.Send(tgMsg)
				if err != nil {
					return fmt.Errorf("api.Send: %s", err.Error())
				}
				bot.updateUserSettings(user)
				err = bot.openRoute(RouteSettings, user)
				if err != nil {
					return fmt.Errorf("openRoute: %s", err.Error())
				}
				return nil
			},
		},
	}
}
