package market

import (
	"github.com/everstake/nebulas-tg-bot/log"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

type (
	Market struct {
		priceNAS   decimal.Decimal
		priceNAX   decimal.Decimal
		trackerNAS Tracker
		trackerNAX Tracker
		mu         *sync.Mutex
	}
	Tracker interface {
		GetPrice() (price decimal.Decimal, err error)
	}
)

func NewMarket() *Market {
	priceNAS := decimal.Zero
	trackerNAS := okex{}
	var err error
	for {
		priceNAS, err = trackerNAS.GetPrice()
		if err != nil {
			log.Error("Market: tracker.GetPrice(nas): %s", err.Error())
			<-time.After(time.Second * 5)
		} else {
			break
		}
	}
	priceNAX := decimal.Zero
	trackerNAX := gate{}
	for {
		priceNAX, err = trackerNAX.GetPrice()
		if err != nil {
			log.Error("Market: tracker.GetPrice(nax): %s", err.Error())
			<-time.After(time.Second * 5)
		} else {
			break
		}
	}
	return &Market{
		priceNAS:   priceNAS,
		priceNAX:   priceNAX,
		trackerNAS: &trackerNAS,
		trackerNAX: &trackerNAX,
		mu:         &sync.Mutex{},
	}
}

func (m *Market) Run() {
	for {
		<-time.After(time.Minute * 5)
		nas, err := m.trackerNAS.GetPrice()
		if err != nil {
			log.Error("Market: tracker.GetPrice(nas): %s", err.Error())
		} else {
			m.mu.Lock()
			m.priceNAS = nas
			m.mu.Unlock()
		}
		nax, err := m.trackerNAX.GetPrice()
		if err != nil {
			log.Error("Market: tracker.GetPrice(nax): %s", err.Error())
		} else {
			m.mu.Lock()
			m.priceNAX = nax
			m.mu.Unlock()
		}
	}
}

func (m *Market) GetNASPrice() decimal.Decimal {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.priceNAS.Add(decimal.Zero)
}

func (m *Market) GetNAXPrice() decimal.Decimal {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.priceNAX.Add(decimal.Zero)
}
