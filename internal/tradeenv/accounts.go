package tradeenv

func (e *TradeEnv) GetUnoccupiedAccount() string {
	e.accountsOccupationRegistry.mu.Lock()
	for _, account := range e.accounts {
		if e.accountsOccupationRegistry.table[account.Id] != true {
			return account.Id
		}
	}
	return ""
}

func (e *TradeEnv) SetAccountOccupied(accountId string) {
	e.accountsOccupationRegistry.table[accountId] = true
	e.accountsOccupationRegistry.mu.Unlock()
}

func (e *TradeEnv) SetAccountUnoccupied(accountId string) {
	e.accountsOccupationRegistry.mu.Lock()
	e.accountsOccupationRegistry.table[accountId] = false
	e.accountsOccupationRegistry.mu.Unlock()
}
