package tradeenv

import (
	"log"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client"
	"tinkoff-invest-contest/internal/config"
	"tinkoff-invest-contest/internal/utils"
)

type TradeEnv struct {
	token     string
	isSandbox bool
	fee       float64

	accountsRegistry *accountsRegistry

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

	tradeEnv.accountsRegistry = newAccountsRegistry()

	if config.IsSandbox {
		for i := 0; i < config.NumAccounts; i++ {
			accountResp, err := tradeEnv.Client.OpenSandboxAccount()
			utils.MaybeCrash(err)
			_, err = tradeEnv.Client.SandboxPayIn(accountResp.AccountId, "rub", config.Money) // TODO: allow to configure sandbox accounts through config file or using api endpoints
			utils.MaybeCrash(err)
			tradeEnv.accountsRegistry.accounts[accountResp.AccountId] = make(map[string]*moneyPosition)
			tradeEnv.accountsRegistry.accounts[accountResp.AccountId]["rub"] = &moneyPosition{
				amount:   config.Money,
				occupied: false,
			}
			tradeEnv.accountsRegistry.accounts[accountResp.AccountId]["usd"] = &moneyPosition{
				amount:   config.Money,
				occupied: false,
			}
		}
	} else {
		accounts, err := tradeEnv.Client.GetAccounts()
		utils.MaybeCrash(err)
		for _, account := range accounts {
			positions, err := tradeEnv.Client.GetPositions(account.Id)
			utils.MaybeCrash(err)
			var rubAmount, usdAmount float64
			for _, position := range positions.Money {
				if position.Currency == "rub" {
					rubAmount = utils.FloatFromMoneyValue(position)
				}
				if position.Currency == "usd" {
					usdAmount = utils.FloatFromMoneyValue(position)
				}
			}
			tradeEnv.accountsRegistry.accounts[account.Id] = make(map[string]*moneyPosition)
			tradeEnv.accountsRegistry.accounts[account.Id]["rub"] = &moneyPosition{
				amount:   rubAmount,
				occupied: false,
			}
			tradeEnv.accountsRegistry.accounts[account.Id]["usd"] = &moneyPosition{
				amount:   usdAmount,
				occupied: false,
			}
		}
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
		for accountId, _ := range e.accountsRegistry.accounts {
			_, err := e.Client.CloseSandboxAccount(accountId)
			if err != nil {
				log.Println(utils.PrettifyError(err))
			}
		}
	}
}
