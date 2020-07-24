package bot

import (
	"fmt"
	"github.com/everstake/nebulas-tg-bot/log"
	"github.com/everstake/nebulas-tg-bot/models"
	"github.com/everstake/nebulas-tg-bot/services/node"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"time"
)

func (bot *Bot) Parsing() {
	for {
		err := func() error {
			state, err := bot.dao.GetState(models.StateCurrentHeight)
			if err != nil {
				return fmt.Errorf("dao.GetState: %s", err.Error())
			}
			currentHeight, err := strconv.ParseUint(state.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("strconv.ParseUint: %s", err.Error())
			}
			currentHeight++
			latestBlock, err := bot.node.GetLatestIrreversibleBlock()
			if err != nil {
				return fmt.Errorf("node.GetLatestIrreversibleBlock: %s", err.Error())
			}
			latestBlockheight := latestBlock.Result.Height
			//latestBlockheight = latestBlockheight - 5 // bug with empty txs from node
			if latestBlockheight <= currentHeight {
				return nil
			}
			for h := currentHeight; h <= latestBlockheight; h++ {
				block, err := bot.node.GetBlock(h)
				if err != nil {
					return fmt.Errorf("node.GetBlock(%d): %s", h, err.Error())
				}
				for _, tx := range block.Result.Transactions {
					if bot.addressExist(tx.From) {
						bot.txNotify(tx.From, tx)
					}
					if bot.addressExist(tx.To) {
						bot.txNotify(tx.To, tx)
					}
				}
				state.Value = fmt.Sprintf("%d", h)
				err = bot.dao.UpdateState(state)
				if err != nil {
					log.Error("Bot: Parsing: dao.UpdateState: %s", err.Error())
				}
			}
			return nil
		}()
		if err != nil {
			log.Error("Bot Parsing: %s", err.Error())
		}
		<-time.After(time.Second * 2)
	}
}

func (bot *Bot) txNotify(address string, tx node.Transaction) {
	status := "success"
	if tx.Status != 1 {
		status = "failed"
	}
	value := tx.Value.Div(node.PrecisionDivNAS)
	txt := fmt.Sprintf(
		"[Transaction]\nHash: %s\nFrom: %s\nTo: %s\nValue: %s\nBlock: %d\nStatus: %s",
		tx.Hash,
		tx.From,
		tx.To,
		value.Truncate(4).String(),
		tx.BlockHeight,
		status,
	)
	users := make(map[uint64]models.User)
	bot.mu.RLock()
	for userID := range bot.addresses[address] {
		user, ok := bot.users[userID]
		if !ok {
			continue
		}
		users[userID] = user
	}
	bot.mu.RUnlock()
	for _, user := range users {
		if user.Mute {
			continue
		}
		if value.GreaterThan(user.MinTreshold) && value.LessThanOrEqual(user.MaxTreshold) {
			msg := tgbotapi.NewMessage(user.TgID, txt)
			_, err := bot.api.Send(msg)
			if err != nil {
				log.Error("Bot: txNotify: %s", err.Error())
			}
		}
	}
}
