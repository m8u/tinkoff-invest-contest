package app

import (
	"tinkoff-invest-contest/internal/bots"
	"tinkoff-invest-contest/internal/tradeenv"
)

var (
	CombatEnv  *tradeenv.TradeEnv
	SandboxEnv *tradeenv.TradeEnv
	Bots       map[string]bots.Bot
)

func Init() {
	Bots = make(map[string]bots.Bot)
}
