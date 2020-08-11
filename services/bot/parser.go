package bot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/everstake/nebulas-tg-bot/log"
	"github.com/everstake/nebulas-tg-bot/models"
	"github.com/everstake/nebulas-tg-bot/services/node"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

const pollingCycleBlocks = 210
const blocksInGovernancePeriod = 820
const candidateNode = 3
const consensusNode = 2
const startPointBlock = 4893100 // 23348  polling cycle
const BlockedByUserErr  = "Forbidden: bot was blocked by the user"

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
				// update nodes list
				if h%10 == 0 { // every 10 blocks
					err = bot.setNodes()
					if err != nil {
						return fmt.Errorf("setNodes: %s", err.Error())
					}
					bot.checkStabilityIndexes()
				}

				if (h-startPointBlock)%(pollingCycleBlocks*blocksInGovernancePeriod) == 0 {
					bot.notifyGovernanceCandidates()
				}

				block, err := bot.node.GetBlock(h)
				if err != nil {
					return fmt.Errorf("node.GetBlock(%d): %s", h, err.Error())
				}
				for _, tx := range block.Result.Transactions {
					if tx.To == StakingContract {
						err = bot.stakingNotify(tx)
						if err != nil {
							return fmt.Errorf("stakingNotify (height: %d): %s", h, err.Error())
						}
						continue
					}
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
		txt := fmt.Sprintf(
			bot.dictionary.Get("t.transaction", user.Lang),
			tx.Hash,
			tx.From,
			tx.To,
			value.Truncate(4).String(),
			tx.BlockHeight,
			status,
		)

		if value.GreaterThan(user.MinThreshold) && value.LessThanOrEqual(user.MaxThreshold) {
			url := fmt.Sprintf("https://explorer.nebulas.io/#/tx/%s", tx.Hash)
			var keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL(bot.dictionary.Get("b.link", user.Lang), url),
				),
			)
			msg := tgbotapi.NewMessage(user.TgID, txt)
			msg.ReplyMarkup = keyboard
			_, err := bot.api.Send(msg)
			if err != nil {
				log.Error("Bot: txNotify: %s", err.Error())
			}
		}
	}
}

func (bot *Bot) stakingNotify(tx node.Transaction) error {
	if tx.Status != 1 {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(tx.Data)
	if err != nil {
		return fmt.Errorf("base64.DecodeString: %s", err.Error())
	}
	var contract node.CallContract
	err = json.Unmarshal(data, &contract)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %s", err.Error())
	}

	switch contract.Function {
	case "cancelVote", "vote":
		var args []string
		err = json.Unmarshal([]byte(contract.Args), &args)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %s", err.Error())
		}
		if len(args) < 2 {
			return nil
		}
		nodeID := args[0]
		bot.mu.RLock()
		validator, ok := bot.nodes[nodeID]
		if !ok {
			log.Warn("Bot: Parser: validator %s not found", nodeID)
			return nil
		}
		value, err := decimal.NewFromString(args[1])
		if err != nil {
			log.Warn("Bot: Parser: decimal.NewFromString: %s", err.Error())
			return nil
		}
		value = value.Div(node.PrecisionDivNAX)
		addresses := getUniqStrings([]string{
			validator.Accounts.ConsensusManager,
			validator.Accounts.GovManager,
			validator.Accounts.Registrant,
			validator.Accounts.StakingAccount,
		})
		users := make(map[uint64]models.User)
		for _, address := range addresses {
			validators, ok := bot.validators[address]
			if !ok {
				continue
			}
			for userID := range validators {
				user, ok := bot.users[userID]
				if ok {
					users[userID] = user
				}
			}
		}
		userIDs, ok := bot.addresses[tx.From]
		if ok {
			for userID := range userIDs {
				user, ok := bot.users[userID]
				if ok {
					users[userID] = user
				}
			}
		}
		bot.mu.RUnlock()
		for _, user := range users {
			if user.Mute {
				continue
			}
			var text string
			if contract.Function == "vote" {
				text = fmt.Sprintf(
					bot.dictionary.Get("t.new_delegation", user.Lang),
					tx.From,
					nodeID,
					value.String(),
				)
			} else {
				text = fmt.Sprintf(
					bot.dictionary.Get("t.new_undelegation", user.Lang),
					tx.From,
					nodeID,
					value.String(),
				)
			}
			msg := tgbotapi.NewMessage(user.TgID, text)
			err := bot.sendMsg(msg)
			if err != nil {
				return fmt.Errorf("api.Send: %s", err.Error())
			}
		}
	case "transfer":
		var args []string
		err = json.Unmarshal([]byte(contract.Args), &args)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %s", err.Error())
		}
		if len(args) < 2 {
			return nil
		}
		to := args[0]
		value, err := decimal.NewFromString(args[1])
		if err != nil {
			log.Warn("Bot: Parser: decimal.NewFromString: %s", err.Error())
			return nil
		}
		value = value.Div(node.PrecisionDivNAX)
		users := make(map[uint64]models.User)
		bot.mu.RLock()
		userIDs, ok := bot.addresses[tx.From]
		if ok {
			for userID := range userIDs {
				user, ok := bot.users[userID]
				if ok {
					users[user.ID] = user
				}
			}
		}
		userIDs, ok = bot.addresses[to]
		if ok {
			for userID := range userIDs {
				user, ok := bot.users[userID]
				if ok {
					users[user.ID] = user
				}
			}
		}
		bot.mu.RUnlock()
		for _, user := range users {
			text := fmt.Sprintf(
				bot.dictionary.Get("t.transfer_nax", user.Lang),
				tx.From,
				to,
				value,
			)
			msg := tgbotapi.NewMessage(user.TgID, text)
			err := bot.sendMsg(msg)
			if err != nil {
				return fmt.Errorf("api.Send: %s", err.Error())
			}
		}
	}
	return nil
}

func (bot *Bot) setNodes() error {
	list, err := bot.node.GetNodesList()
	if err != nil {
		return fmt.Errorf("node.GetNodesList: %s", err.Error())
	}
	for _, n := range list {
		bot.nodes[n.ID] = n
	}
	return nil
}

func (bot *Bot) checkStabilityIndexes() {
	messages := make(map[int64]string)
	bot.mu.RLock()
	for _, n := range bot.nodes {
		if n.StabilityIndex != 1 {
			prev, ok := bot.lastStabilityIndexes[n.ID]
			if !ok {
				continue
			}
			if n.StabilityIndex < prev {
				addresses := getUniqStrings([]string{
					n.Accounts.ConsensusManager,
					n.Accounts.GovManager,
					n.Accounts.Registrant,
					n.Accounts.StakingAccount,
				})
				for _, address := range addresses {
					usersIDs, ok := bot.validators[address]
					if !ok {
						continue
					}
					for userID := range usersIDs {
						user, ok := bot.users[userID]
						if !ok {
							continue
						}
						if user.Mute {
							continue
						}
						text := fmt.Sprintf(bot.dictionary.Get("t.changed_stability_index", user.Lang), n.ID, n.StabilityIndex)
						messages[user.TgID] = text
					}
				}
			}
		}
	}
	for _, n := range bot.nodes {
		bot.lastStabilityIndexes[n.ID] = n.StabilityIndex
	}
	bot.mu.RUnlock()

	// send messages
	for tgID, text := range messages {
		msg := tgbotapi.NewMessage(tgID, text)
		_, err := bot.api.Send(msg)
		if err != nil {
			log.Error("Bot: checkStabilityIndexes: api.Send: %s", err.Error())
		}
	}
}

func (bot *Bot) notifyGovernanceCandidates() {
	messages := make(map[int64]string)
	bot.mu.RLock()
	for _, n := range bot.nodes {
		if n.Type == consensusNode || n.Type == candidateNode {
			addresses := getUniqStrings([]string{
				n.Accounts.ConsensusManager,
				n.Accounts.GovManager,
				n.Accounts.Registrant,
				n.Accounts.StakingAccount,
			})
			for _, address := range addresses {
				usersIDs, ok := bot.validators[address]
				if !ok {
					continue
				}
				for userID := range usersIDs {
					user, ok := bot.users[userID]
					if !ok {
						continue
					}
					if user.Mute {
						continue
					}
					text := fmt.Sprintf(bot.dictionary.Get("t.inclusion_governance", user.Lang), n.ID)
					messages[user.TgID] = text
				}
			}
		}
	}
	bot.mu.RUnlock()
	// send messages
	for tgID, text := range messages {
		msg := tgbotapi.NewMessage(tgID, text)
		_, err := bot.api.Send(msg)
		if err != nil {
			log.Error("Bot: notifyGovernanceCandidates: api.Send: %s", err.Error())
		}
	}
}
