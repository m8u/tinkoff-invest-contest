package tradeenv

import (
	"testing"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

func TestTradeEnv_CalculateLotsCanAfford(t *testing.T) {
	e := New(utils.GetSandboxToken(), true)
	type args struct {
		direction      investapi.OrderDirection
		maxDealValue   float64
		figi           string
		instrumentType utils.InstrumentType
		price          *investapi.Quotation
		fee            float64
	}
	tests := []struct {
		name              string
		createAccountArgs []map[string]float64
		args              args
		want              int64
	}{
		{
			name: "test1",
			createAccountArgs: []map[string]float64{
				{"rub": 10000, "usd": 0},
			},
			args: args{
				direction:      investapi.OrderDirection_ORDER_DIRECTION_BUY,
				maxDealValue:   10000,
				figi:           "BBG006L8G4H1",
				instrumentType: utils.InstrumentType_INSTRUMENT_TYPE_SHARE,
				price:          utils.FloatToQuotation(1000),
				fee:            0,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instrument, _ := e.Client.InstrumentByFigi(tt.args.figi, tt.args.instrumentType)
			if got := e.CalculateLotsCanAfford(tt.args.direction, tt.args.maxDealValue, instrument, tt.args.price, tt.args.fee); got != tt.want {
				t.Errorf("CalculateLotsCanAfford() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTradeEnv_CalculateMaxDealValue(t *testing.T) {
	e := New(utils.GetSandboxToken(), true)
	e.CreateSandboxAccount(map[string]float64{"rub": 10000, "usd": 0})

	type args struct {
		direction      investapi.OrderDirection
		figi           string
		instrumentType utils.InstrumentType
		price          *investapi.Quotation
		allowMargin    bool
	}
	tests := []struct {
		name              string
		createAccountArgs []map[string]float64
		args              args
		want              float64
		wantErr           bool
	}{
		{
			name: "test1",
			args: args{
				direction:      investapi.OrderDirection_ORDER_DIRECTION_BUY,
				figi:           "BBG006L8G4H1",
				instrumentType: utils.InstrumentType_INSTRUMENT_TYPE_SHARE,
				price:          utils.FloatToQuotation(1000),
				allowMargin:    false,
			},
			want:    10000,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instrument, _ := e.Client.InstrumentByFigi(tt.args.figi, tt.args.instrumentType)
			accountId, unlock, _ := e.GetUnoccupiedAccount(instrument.GetCurrency())
			got := e.CalculateMaxDealValue(accountId, tt.args.direction, instrument, tt.args.price, tt.args.allowMargin)
			if got != tt.want {
				t.Errorf("CalculateMaxDealValue() got = %v, want %v", got, tt.want)
			}
			unlock()
		})
	}
	for id := range e.accounts {
		_, _ = e.Client.CloseSandboxAccount(id)
	}
}

func TestTradeEnv_GetLotsHave(t *testing.T) {
	e := New(utils.GetSandboxToken(), true)
	e.CreateSandboxAccount(map[string]float64{"rub": 100000, "usd": 0})
	type args struct {
		figi           string
		instrumentType utils.InstrumentType
	}
	tests := []struct {
		name     string
		args     args
		wantLots int64
		wantErr  bool
	}{
		{
			name: "test1",
			args: args{
				figi:           "BBG006L8G4H1",
				instrumentType: utils.InstrumentType_INSTRUMENT_TYPE_SHARE,
			},
			wantLots: 1,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instrument, _ := e.Client.InstrumentByFigi(tt.args.figi, tt.args.instrumentType)
			accountId, unlock, _ := e.GetUnoccupiedAccount(instrument.GetCurrency())
			_ = e.DoOrder(tt.args.figi, tt.wantLots, utils.FloatToQuotation(1000),
				investapi.OrderDirection_ORDER_DIRECTION_BUY, accountId, investapi.OrderType_ORDER_TYPE_MARKET)
			gotLots, err := e.GetLotsHave(accountId, instrument)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLotsHave() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLots != tt.wantLots {
				t.Errorf("GetLotsHave() gotLots = %v, want %v", gotLots, tt.wantLots)
			}
			unlock()
		})
	}
	for id := range e.accounts {
		_, _ = e.Client.CloseSandboxAccount(id)
	}
}
