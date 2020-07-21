package market

import (
	"github.com/everstake/nebulas-tg-bot/log"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

type (
	Market struct {
		price   decimal.Decimal
		tracker Tracker
		mu      *sync.Mutex
	}
	Tracker interface {
		GetPrice() (price decimal.Decimal, err error)
	}
)

func NewMarket() *Market {
	price := decimal.Zero
	tracker := okex{}
	var err error
	for {
		price, err = tracker.GetPrice()
		if err != nil {
			log.Error("Market: tracker.GetPrice: %s", err.Error())
			<-time.After(time.Second * 5)
		} else {
			break
		}
	}
	return &Market{
		price:   price,
		tracker: &tracker,
		mu:      &sync.Mutex{},
	}
}

func (m *Market) Run() {
	var err error
	for {
		<-time.After(time.Minute * 5)
		m.mu.Lock()
		m.price, err = m.tracker.GetPrice()
		m.mu.Unlock()
		if err != nil {
			log.Error("Market: tracker.GetPrice: %s", err.Error())
		}
	}
}

func (m *Market) GetPrice() decimal.Decimal {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.price.Add(decimal.Zero)
}
