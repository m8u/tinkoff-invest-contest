package app

import (
	"sync"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/bot"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"

	// Shadow-import your strategies here
	_ "tinkoff-invest-contest/internal/strategies/bollinger"
	_ "tinkoff-invest-contest/internal/strategies/consecutive_ratio"
)

type botsTable struct {
	Lock  sync.Mutex
	Table map[string]*bot.Bot
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
		Table: make(map[string]*bot.Bot),
	}
}
