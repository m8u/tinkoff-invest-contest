package tradeenv

import "sync"

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
	for accountId, moneyPositions := range e.accountsRegistry.accounts {
		if !(moneyPositions[currency].occupied) {
			return accountId, unlock
		}
	}
	return "", unlock
}

func (e *TradeEnv) SetAccountOccupied(accountId string, currency string) {
	e.accountsRegistry.accounts[accountId][currency].occupied = true
}

func (e *TradeEnv) SetAccountUnoccupied(accountId string, currency string) {
	e.accountsRegistry.mu.Lock()
	e.accountsRegistry.accounts[accountId][currency].occupied = false
	e.accountsRegistry.mu.Unlock()
}
