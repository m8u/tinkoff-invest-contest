package bollinger

import (
	"reflect"
	"testing"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/strategies"
	indicators "tinkoff-invest-contest/internal/technical_indicators"
	"tinkoff-invest-contest/internal/utils"
)

func Test_bollinger_GetTradeSignal(t *testing.T) {
	type fields struct {
		indicator      *indicators.BollingerBands
		pointDeviation float64
	}
	type args struct {
		candles []*investapi.HistoricCandle
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantSignal *utils.TradeSignal
		wantValues map[string]any
	}{
		{
			name: "test1",
			fields: fields{
				indicator:      indicators.NewBollingerBands(3),
				pointDeviation: 0.01,
			},
			args: args{
				candles: []*investapi.HistoricCandle{
					{
						High:  utils.FloatToQuotation(186.89),
						Low:   utils.FloatToQuotation(186.22),
						Close: utils.FloatToQuotation(186.45),
					},
					{
						High:  utils.FloatToQuotation(186.49),
						Low:   utils.FloatToQuotation(185.87),
						Close: utils.FloatToQuotation(185.13),
					},
					{
						High:  utils.FloatToQuotation(186.19),
						Low:   utils.FloatToQuotation(184.95),
						Close: utils.FloatToQuotation(184.95),
					},
				},
			},
			wantSignal: &utils.TradeSignal{
				Direction: investapi.OrderDirection_ORDER_DIRECTION_BUY,
			},
			wantValues: map[string]any{
				"bollinger_lower_bound": 182.68758202226874,
				"bollinger_upper_bound": 187.97908464439794,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bollingerStrategy{
				indicator:      tt.fields.indicator,
				pointDeviation: tt.fields.pointDeviation,
			}
			gotSignal, gotValues := b.GetTradeSignal(strategies.MarketData{
				Candles: tt.args.candles,
			})
			if !reflect.DeepEqual(gotSignal, tt.wantSignal) {
				t.Errorf("GetTradeSignal() gotSignal = %v, want %v", gotSignal, tt.wantSignal)
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("GetTradeSignal() gotValues = %v, want %v", gotValues, tt.wantValues)
			}
		})
	}
}
