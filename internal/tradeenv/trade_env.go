package tradeenv

import (
	"log"
	"sync"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client"
	"tinkoff-invest-contest/internal/utils"
)

type TradeEnv struct {
	token     string
	isSandbox bool
	CombatFee float64

	Mu            sync.RWMutex
	accounts      map[string]map[string]*moneyPosition
	subscriptions *subscriptions
	Channels      map[string]*MarketDataChannelStack

	Client *client.Client
}

func New(token string, isSandbox bool) *TradeEnv {
	tradeEnv := &TradeEnv{
		token:         token,
		isSandbox:     isSandbox,
		accounts:      make(map[string]map[string]*moneyPosition),
		subscriptions: new(subscriptions),
		Channels:      make(map[string]*MarketDataChannelStack),
		Client:        client.NewClient(token),
	}
	tradeEnv.Client.InitMarketDataStream()

	if !isSandbox {
		tradeEnv.loadCombatAccounts()

		info, err := tradeEnv.Client.GetInfo()
		utils.MaybeCrash(err)
		tradeEnv.CombatFee = utils.Fees[utils.Tariff(info.Tariff)]
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
