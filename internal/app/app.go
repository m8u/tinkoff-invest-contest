package app

import (
	"sync"
	"tinkoff-invest-contest/internal/bots"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

type botsTable struct {
	Lock  sync.Mutex
	Table map[string]bots.Bot
}

var (
	SandboxEnv *tradeenv.TradeEnv
	CombatEnv  *tradeenv.TradeEnv
	Bots       *botsTable
)

func init() {
	SandboxEnv = tradeenv.New(utils.GetSandboxToken(), true, utils.Fees[utils.Trader])
	Bots = &botsTable{
		Table: make(map[string]bots.Bot),
	}
}
