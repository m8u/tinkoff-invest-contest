package tradeenv

import (
	"log"
	"sort"
	"strings"
	"sync"
	"tinkoff-invest-contest/internal/utils"
)

type moneyPosition struct {
	amount   float64
	occupied bool
}

type accountsRegistry struct {
	mu       sync.Mutex
	accounts map[string]map[string]*moneyPosition
}

func newAccountsRegistry() (registry *accountsRegistry) {
	registry = &accountsRegistry{
		mu:       sync.Mutex{},
		accounts: make(map[string]map[string]*moneyPosition),
	}
	return
}

// GetUnoccupiedAccount returns an unnocupied account id or an empty string, if there isn't any.
// Remember to call unlock() after calling SetAccountOccupied()
// or after right calling this function, if obtained account does not satisfy fund requirements or no account returned
func (e *TradeEnv) GetUnoccupiedAccount(currency string) (accountId string, unlock func()) {
	unlock = func() {
		e.accountsRegistry.mu.Unlock()
	}
	e.accountsRegistry.mu.Lock()
	var maxMoneyAmount float64
	for id, moneyPositions := range e.accountsRegistry.accounts {
		if !(moneyPositions[currency].occupied) && moneyPositions[currency].amount > maxMoneyAmount {
			maxMoneyAmount = moneyPositions[currency].amount
			accountId = id
		}
	}
	return
}

func (e *TradeEnv) SetAccountOccupied(accountId string, currency string) {
	e.accountsRegistry.accounts[accountId][currency].occupied = true
}

func (e *TradeEnv) SetAccountUnoccupied(accountId string, currency string) {
	positions, err := e.Client.WrapGetPositions(e.isSandbox, accountId)
	utils.MaybeCrash(err)
	e.accountsRegistry.mu.Lock()
	for _, moneyPosition := range positions.Money {
		e.accountsRegistry.accounts[accountId][moneyPosition.Currency].amount = utils.MoneyValueToFloat(moneyPosition)
	}
	e.accountsRegistry.accounts[accountId][currency].occupied = false
	e.accountsRegistry.mu.Unlock()
}

func (e *TradeEnv) CreateSandboxAccount(money map[string]float64) (accountId string) {
	accountResp, err := e.Client.OpenSandboxAccount()
	utils.MaybeCrash(err)
	accountId = accountResp.AccountId
	e.accountsRegistry.accounts[accountResp.AccountId] = make(map[string]*moneyPosition)
	for currency, amount := range money {
		if amount > 0 {
			_, err = e.Client.SandboxPayIn(accountResp.AccountId, currency, amount)
			utils.MaybeCrash(err)
		}
		e.accountsRegistry.accounts[accountResp.AccountId][currency] = &moneyPosition{
			amount:   amount,
			occupied: false,
		}
	}
	return
}

func (e *TradeEnv) RemoveSandboxAccount(id string) {
	if _, ok := e.accountsRegistry.accounts[id]; ok {
		e.accountsRegistry.mu.Lock()
		delete(e.accountsRegistry.accounts, id)
		e.accountsRegistry.mu.Unlock()

		_, _ = e.Client.CloseSandboxAccount(id)
	}
}

func (e *TradeEnv) loadCombatAccounts() {
	accounts, err := e.Client.GetAccounts()
	utils.MaybeCrash(err)
	for _, account := range accounts {
		positions, err := e.Client.GetPositions(account.Id)
		utils.MaybeCrash(err)
		e.accountsRegistry.accounts[account.Id] = make(map[string]*moneyPosition)
		for _, position := range positions.Money {
			log.Println(utils.MoneyValueToFloat(position))
			e.accountsRegistry.accounts[account.Id][position.Currency] = &moneyPosition{
				amount:   utils.MoneyValueToFloat(position),
				occupied: false,
			}
		}
	}
}

type accountsPayloadEntry struct {
	Id          string  `json:"id"`
	RUBAmount   float64 `json:"rubAmount"`
	USDAmount   float64 `json:"usdAmount"`
	RUBOccupied bool    `json:"rubOccupied"`
	USDOccupied bool    `json:"usdOccupied"`
}

func (e *TradeEnv) GetAccountsPayload() any {
	var accounts []accountsPayloadEntry
	for id, account := range e.accountsRegistry.accounts {
		rubPosition, ok := account["rub"]
		if !ok {
			rubPosition = &moneyPosition{
				amount:   0,
				occupied: false,
			}
		}
		usdPosition, ok := account["usd"]
		if !ok {
			usdPosition = &moneyPosition{
				amount:   0,
				occupied: false,
			}
		}
		accounts = append(accounts, accountsPayloadEntry{
			Id:          id,
			RUBAmount:   rubPosition.amount,
			USDAmount:   usdPosition.amount,
			RUBOccupied: rubPosition.occupied,
			USDOccupied: usdPosition.occupied,
		})
	}
	sort.Slice(accounts, func(i, j int) bool {
		return strings.Compare(accounts[i].Id, accounts[j].Id) == -1
	})
	return accounts
}
