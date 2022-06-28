package tradeenv

// GetUnoccupiedAccount returns an unnocupied account id or an empty string, if there isn't any.
// Remember to call unlock() after calling SetAccountOccupied()
// or after right calling this function, if obtained account does not satisfy fund requirements or no account returned
func (e *TradeEnv) GetUnoccupiedAccount() (accountId string, unlock func()) {
	unlock = func() {
		e.accountsOccupationRegistry.mu.Unlock()
	}
	e.accountsOccupationRegistry.mu.Lock()
	for _, account := range e.accounts {
		if e.accountsOccupationRegistry.table[account.Id] != true {
			return account.Id, unlock
		}
	}
	return "", unlock
}

func (e *TradeEnv) SetAccountOccupied(accountId string) {
	e.accountsOccupationRegistry.table[accountId] = true
}

func (e *TradeEnv) SetAccountUnoccupied(accountId string) {
	e.accountsOccupationRegistry.mu.Lock()
	e.accountsOccupationRegistry.table[accountId] = false
	e.accountsOccupationRegistry.mu.Unlock()
}
