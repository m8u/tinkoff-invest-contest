package indicators

import (
	"testing"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

func TestBollingerBands_Calculate(t *testing.T) {
	type fields struct {
		coef float64
	}
	type args struct {
		candles []*investapi.HistoricCandle
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantLowerBound float64
		wantUpperBound float64
	}{
		{
			name: "test1",
			fields: fields{
				coef: 3,
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
			wantLowerBound: 182.68758202226874,
			wantUpperBound: 187.97908464439794,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bollinger := &BollingerBands{
				coef: tt.fields.coef,
			}
			gotLowerBound, gotUpperBound := bollinger.Calculate(tt.args.candles)
			if gotLowerBound != tt.wantLowerBound {
				t.Errorf("Calculate() gotLowerBound = %v, want %v", gotLowerBound, tt.wantLowerBound)
			}
			if gotUpperBound != tt.wantUpperBound {
				t.Errorf("Calculate() gotUpperBound = %v, want %v", gotUpperBound, tt.wantUpperBound)
			}
		})
	}
}
