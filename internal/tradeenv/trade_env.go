package tradeenv

import (
	"log"
	"sync"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type TradeEnv struct {
	token     string
	isSandbox bool
	Fee       float64

	mu            sync.RWMutex
	accounts      map[string]map[string]*moneyPosition
	subscriptions *subscriptions
	marketData    map[string]*MarketDataChannelStack
	trades        map[string]chan *investapi.OrderTrades

	Client *client.Client
}

func New(token string, isSandbox bool) *TradeEnv {
	tradeEnv := &TradeEnv{
		token:         token,
		isSandbox:     isSandbox,
		accounts:      make(map[string]map[string]*moneyPosition),
		subscriptions: new(subscriptions),
		marketData:    make(map[string]*MarketDataChannelStack),
		trades:        make(map[string]chan *investapi.OrderTrades),
		Client:        client.NewClient(token),
	}
	tradeEnv.Client.InitMarketDataStream()

	if !isSandbox {
		tradeEnv.loadCombatAccounts()
		go tradeEnv.Client.RunTradesStreamLoop(tradeEnv.handleTradesStream)

		info, err := tradeEnv.Client.GetInfo()
		utils.MaybeCrash(err)
		tradeEnv.Fee = utils.Fees[utils.Tariff(info.Tariff)]
	} else {
		tradeEnv.Fee = 0
	}

	go tradeEnv.Client.RunMarketDataStreamLoop(tradeEnv.handleMarketDataStream, tradeEnv.handleResubscribe)

	go func() {
		appstate.ExitActionsWG.Wait()
		tradeEnv.exitActions()
	}()

	appstate.PostExitActionsWG.Add(1)

	return tradeEnv
}

func (e *TradeEnv) exitActions() {
	defer appstate.PostExitActionsWG.Done()

	if e.isSandbox {
		for accountId := range e.accounts {
			_, err := e.Client.CloseSandboxAccount(accountId)
			if err != nil {
				log.Println(utils.PrettifyError(err))
			}
		}
	}
}
