package bot

import (
	"fmt"
	"github.com/everstake/nebulas-tg-bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	RouteStart      = "start"
	RouteChooseLang = "choose_lang"
	RouteSettings   = "settings"
)

type Route struct {
	Request  func(user models.User) error
	Response func(update tgbotapi.Update, user models.User) error
}

func (bot *Bot) SetRoutes() {
	bot.routes = map[string]Route{
		RouteChooseLang: {
			Request: func(user models.User) error {
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
			Response: func(update tgbotapi.Update, user models.User) error {
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
			Request: func(user models.User) error {
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
			Response: func(update tgbotapi.Update, user models.User) error {
				switch update.Message.Text {
				case bot.dictionary.Get("b.settings", user.Lang):
					err := bot.openRoute(RouteSettings, user)
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
			Request: func(user models.User) error {
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
			Response: func(update tgbotapi.Update, user models.User) error {
				switch update.Message.Text {
				case bot.dictionary.Get("b.change_lang", user.Lang):
					err := bot.openRoute("choose_lang", user)
					if err != nil {
						return fmt.Errorf("openRoute: %s", err.Error())
					}
				case bot.dictionary.Get("b.return_back", user.Lang):
					err := bot.openRoute("start", user)
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
	}
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
	err = route.Request(user)
	if err != nil {
		return fmt.Errorf("actions(%s): %s", key, err.Error())
	}
	return nil
}
