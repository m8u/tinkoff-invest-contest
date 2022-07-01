package tradeenv

import (
	"log"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client"
	"tinkoff-invest-contest/internal/utils"
)

type TradeEnv struct {
	token     string
	isSandbox bool

	accountsRegistry *accountsRegistry

	Client   *client.Client
	Channels map[string]MarketDataChannelStack
}

func New(token string, isSandbox bool) *TradeEnv {
	tradeEnv := &TradeEnv{
		token:            token,
		isSandbox:        isSandbox,
		accountsRegistry: newAccountsRegistry(),
		Client:           client.NewClient(token),
		Channels:         make(map[string]MarketDataChannelStack),
	}
	tradeEnv.Client.InitMarketDataStream()

	if !isSandbox {
		tradeEnv.loadCombatAccounts()
	}

	go tradeEnv.Client.RunMarketDataStreamLoop(tradeEnv.handleMarketDataStream)

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
		for accountId := range e.accountsRegistry.accounts {
			_, err := e.Client.CloseSandboxAccount(accountId)
			if err != nil {
				log.Println(utils.PrettifyError(err))
			}
		}
	}
}
