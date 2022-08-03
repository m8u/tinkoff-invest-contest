package app

import (
	"sync"
	"tinkoff-invest-contest/internal/appstate"
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
	appstate.ExitActionsWG.Add(1)
	SandboxEnv = tradeenv.New(utils.GetSandboxToken(), true)
	CombatEnv = tradeenv.New(utils.GetCombatToken(), false)
	Bots = &botsTable{
		Table: make(map[string]bots.Bot),
	}
}
