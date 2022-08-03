package tradeenv

import (
	"sort"
	"strings"
	"tinkoff-invest-contest/internal/utils"
)

type moneyPosition struct {
	amount   float64
	occupied bool
}

// GetUnoccupiedAccount returns an unnocupied account id with the highest amount of requested currency,
// or an empty string, if there isn't any.
// If you're not going to use obtained account for some reason, make sure to call discard().
// Call unlock() after getting all necessary info and deciding to use account or not.
// If no account returned - calling any of these is not needed.
func (e *TradeEnv) GetUnoccupiedAccount(currency string) (accountId string, discard func(), unlock func()) {
	unlock = func() {
		e.mu.Unlock()
	}
	discard = func() {
		e.accounts[accountId][currency].occupied = false
	}
	e.mu.Lock()
	var maxMoneyAmount float64
	for id, moneyPositions := range e.accounts {
		if !(moneyPositions[currency].occupied) && moneyPositions[currency].amount > maxMoneyAmount {
			maxMoneyAmount = moneyPositions[currency].amount
			accountId = id
		}
	}
	if accountId == "" {
		unlock()
		return
	}
	e.accounts[accountId][currency].occupied = true
	return
}

func (e *TradeEnv) ReleaseAccount(accountId string, currency string) {
	positions, err := e.Client.WrapGetPositions(e.isSandbox, accountId)
	utils.MaybeCrash(err)
	e.mu.Lock()
	for _, moneyPosition := range positions.Money {
		e.accounts[accountId][moneyPosition.Currency].amount = utils.MoneyValueToFloat(moneyPosition)
	}
	e.accounts[accountId][currency].occupied = false
	e.mu.Unlock()
}

func (e *TradeEnv) CreateSandboxAccount(money map[string]float64) (accountId string) {
	accountResp, err := e.Client.OpenSandboxAccount()
	utils.MaybeCrash(err)
	accountId = accountResp.AccountId
	e.mu.Lock()
	e.accounts[accountResp.AccountId] = make(map[string]*moneyPosition)
	for currency, amount := range money {
		if amount > 0 {
			_, err = e.Client.SandboxPayIn(accountResp.AccountId, currency, amount)
			utils.MaybeCrash(err)
		}
		e.accounts[accountResp.AccountId][currency] = &moneyPosition{
			amount:   amount,
			occupied: false,
		}
	}
	e.mu.Unlock()
	return
}

func (e *TradeEnv) RemoveSandboxAccount(id string) {
	if _, ok := e.accounts[id]; ok {
		e.mu.Lock()
		delete(e.accounts, id)
		e.mu.Unlock()

		_, _ = e.Client.CloseSandboxAccount(id)
	}
}

func (e *TradeEnv) loadCombatAccounts() {
	accounts, err := e.Client.GetAccounts()
	utils.MaybeCrash(err)
	e.mu.Lock()
	for _, account := range accounts {
		positions, err := e.Client.GetPositions(account.Id)
		utils.MaybeCrash(err)
		e.accounts[account.Id] = make(map[string]*moneyPosition)
		for _, position := range positions.Money {
			e.accounts[account.Id][position.Currency] = &moneyPosition{
				amount:   utils.MoneyValueToFloat(position),
				occupied: false,
			}
		}
	}
	e.mu.Unlock()
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
	for id, account := range e.accounts {
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
