package tradeenv

import (
	"testing"
	"tinkoff-invest-contest/internal/utils"
)

func TestTradeEnv_CreateSandboxAccount(t *testing.T) {
	e := New(utils.GetSandboxToken(), true)
	type args struct {
		money map[string]float64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				money: map[string]float64{"rub": 1000, "usd": 123},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e.CreateSandboxAccount(tt.args.money)
			for id, positions1 := range e.accountsRegistry.accounts {
				positions2, err := e.Client.GetSandboxPositions(id)
				if err != nil {
					t.Fatal(err)
				}
				for _, moneyPos := range positions2.Money {
					if moneyPos.Currency == "rub" && positions1["rub"].amount != utils.FloatFromMoneyValue(moneyPos) {
						t.Errorf("CreateSandboxAccount() got accountRegistry rub: %v != GetSandboxPositions rub: %v",
							positions1["rub"].amount, utils.FloatFromMoneyValue(moneyPos))
					}
					if moneyPos.Currency == "usd" && positions1["usd"].amount != utils.FloatFromMoneyValue(moneyPos) {
						t.Errorf("CreateSandboxAccount() got accountRegistry usd: %v != GetSandboxPositions usd: %v",
							positions1["usd"].amount, utils.FloatFromMoneyValue(moneyPos))
					}
				}
			}
			for id := range e.accountsRegistry.accounts {
				_, _ = e.Client.CloseSandboxAccount(id)
			}
		})
	}
}

func TestTradeEnv_GetUnoccupiedAccount(t *testing.T) {
	e := New(utils.GetSandboxToken(), true)
	type args struct {
		currency string
	}
	tests := []struct {
		name              string
		args              args
		createAccountArgs []map[string]float64
	}{
		{
			name: "test1",
			args: args{
				currency: "rub",
			},
			createAccountArgs: []map[string]float64{
				{"rub": 10000, "usd": 0},
				{"rub": 12000, "usd": 0},
				{"rub": 9000, "usd": 20000},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var maxMoney float64
			for _, money := range tt.createAccountArgs {
				if money[tt.args.currency] > maxMoney {
					maxMoney = money[tt.args.currency]
				}
				e.CreateSandboxAccount(money)
			}

			gotAccountId, unlock := e.GetUnoccupiedAccount(tt.args.currency)
			if e.accountsRegistry.accounts[gotAccountId][tt.args.currency].amount != maxMoney {
				t.Errorf("GetUnoccupiedAccount() got %v = %v, want %v = %v",
					tt.args.currency, e.accountsRegistry.accounts[gotAccountId][tt.args.currency].amount,
					tt.args.currency, maxMoney)
			}
			unlock()
			for id := range e.accountsRegistry.accounts {
				_, _ = e.Client.CloseSandboxAccount(id)
			}
		})
	}
}
