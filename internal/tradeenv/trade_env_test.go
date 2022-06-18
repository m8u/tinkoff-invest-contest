package tradeenv

import (
	"github.com/joho/godotenv"
	"os"
	"testing"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/config"
	"tinkoff-invest-contest/internal/utils"
)

func TestTradeEnv_GetAtLeastNLastCandles(t *testing.T) {
	type args struct {
		figi           string
		candleInterval investapi.CandleInterval
		n              int
	}
	_ = godotenv.Load(".env")
	token := os.Getenv("SANDBOX_TOKEN")
	utils.EnsureTinkoffTokenIsProvided(token, true)
	tradeEnv := New(config.Config{
		IsSandbox:   true,
		Token:       token,
		NumAccounts: 0,
		Money:       0,
		Fee:         0,
	})
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				figi:           "BBG006L8G4H1",
				candleInterval: investapi.CandleInterval_CANDLE_INTERVAL_1_MIN,
				n:              100,
			},
			want:    100,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tradeEnv.GetAtLeastNLastCandles(tt.args.figi, tt.args.candleInterval, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAtLeastNLastCandles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < tt.want {
				t.Errorf("GetAtLeastNLastCandles() got len = %v, want >= %v", len(got), tt.want)
			}
		})
	}
}
