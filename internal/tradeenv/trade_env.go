package tradeenv

import (
	"log"
	"sync"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/config"
	"tinkoff-invest-contest/internal/utils"
)

type occupationRegistry struct {
	mu    sync.Mutex
	table map[string]bool // {accountId: isOccupied, ...}
}

type TradeEnv struct {
	token                      string
	isSandbox                  bool
	accounts                   []*investapi.Account
	accountsOccupationRegistry *occupationRegistry
	fee                        float64

	Client   *client.Client
	Channels map[string]MarketDataChannelStack
}

func New(config config.Config) *TradeEnv {
	tradeEnv := new(TradeEnv)
	tradeEnv.Client = client.NewClient(config.Token)
	tradeEnv.Client.InitMarketDataStream()
	tradeEnv.isSandbox = true
	tradeEnv.fee = config.Fee
	tradeEnv.Channels = make(map[string]MarketDataChannelStack)

	if config.IsSandbox {
		tradeEnv.accounts = make([]*investapi.Account, 0)
		tradeEnv.accountsOccupationRegistry = new(occupationRegistry)
		tradeEnv.accountsOccupationRegistry.table = make(map[string]bool)
		for i := 0; i < config.NumAccounts; i++ {
			accountResp, err := tradeEnv.Client.OpenSandboxAccount()
			utils.MaybeCrash(err)
			_, err = tradeEnv.Client.SandboxPayIn(accountResp.AccountId, "rub", config.Money) // TODO: allow to configure sandbox accounts through config file or using api endpoints
			utils.MaybeCrash(err)

			tradeEnv.accounts = append(tradeEnv.accounts, &investapi.Account{Id: accountResp.AccountId})
			tradeEnv.accountsOccupationRegistry.table[accountResp.AccountId] = false
		}
	} else {
		accounts, err := tradeEnv.Client.GetAccounts()
		utils.MaybeCrash(err)
		tradeEnv.accounts = accounts
	}

	tradeEnv.token = config.Token

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
		for _, account := range e.accounts {
			_, err := e.Client.CloseSandboxAccount(account.Id)
			if err != nil {
				log.Println(utils.PrettifyError(err))
			}
		}
	}
}
