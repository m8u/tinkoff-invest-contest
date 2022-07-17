package app

import (
	"tinkoff-invest-contest/internal/bots"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

var (
	CombatEnv  *tradeenv.TradeEnv
	SandboxEnv *tradeenv.TradeEnv
	Bots       map[string]bots.Bot
)

func init() {
	SandboxEnv = tradeenv.New(utils.GetSandboxToken(), true, utils.Fees[utils.Trader])
	Bots = make(map[string]bots.Bot)
}
