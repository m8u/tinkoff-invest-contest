package tradeenv

import (
	"testing"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

func TestTradeEnv_DoOrder(t *testing.T) {
	e := New(utils.GetSandboxToken(), true)
	type args struct {
		figi           string
		instrumentType utils.InstrumentType
		quantity       int64
		price          *investapi.Quotation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				figi:           "BBG006L8G4H1",
				instrumentType: utils.InstrumentType_INSTRUMENT_TYPE_SHARE,
				quantity:       1,
				price:          utils.FloatToQuotation(1000),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e.CreateSandboxAccount(map[string]float64{"rub": 100000, "usd": 10000})
			instrument, _ := e.Client.InstrumentByFigi(tt.args.figi, tt.args.instrumentType)
			accountId, unlock, _ := e.GetUnoccupiedAccount(instrument.GetCurrency())
			err := e.DoOrder(tt.args.figi, tt.args.quantity, tt.args.price,
				investapi.OrderDirection_ORDER_DIRECTION_BUY, accountId, investapi.OrderType_ORDER_TYPE_MARKET)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoOrder() (buy) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotLotsHave, _ := e.GetLotsHave(accountId, instrument)
			if gotLotsHave != tt.args.quantity {
				t.Errorf("DoOrder() (buy) gotLotsHave = %v, want %v", gotLotsHave, tt.args.quantity)
				return
			}
			err = e.DoOrder(tt.args.figi, tt.args.quantity, tt.args.price,
				investapi.OrderDirection_ORDER_DIRECTION_SELL, accountId, investapi.OrderType_ORDER_TYPE_MARKET)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoOrder() (sell) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotLotsHave, _ = e.GetLotsHave(accountId, instrument)
			if gotLotsHave != 0 {
				t.Errorf("DoOrder() (sell) gotLotsHave = %v, want %v", gotLotsHave, 0)
				return
			}
			unlock()
			_, _ = e.Client.CloseSandboxAccount(accountId)
		})
	}
}
