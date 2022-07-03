package tradeenv

import (
	"testing"
	"time"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

func TestTradeEnv_GetCandlesFor1NthDayBeforeNow(t *testing.T) {
	e := New(utils.GetSandboxToken(), true)
	type args struct {
		figi           string
		candleInterval investapi.CandleInterval
		n              int
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				figi:           "BBG006L8G4H1",
				candleInterval: investapi.CandleInterval_CANDLE_INTERVAL_1_MIN,
				n:              5,
			},
			want:    time.Now().UTC().Add(-5 * 24 * time.Hour),
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				figi:           "BBG006L8G4H1",
				candleInterval: investapi.CandleInterval_CANDLE_INTERVAL_5_MIN,
				n:              100,
			},
			want:    time.Now().UTC().Add(-100 * 24 * time.Hour),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.GetCandlesFor1NthDayBeforeNow(tt.args.figi, tt.args.candleInterval, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCandlesFor1NthDayBeforeNow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, candle := range got {
				gotYearDay, wantYearDay := candle.Time.AsTime().YearDay(), tt.want.YearDay()
				if !(gotYearDay >= wantYearDay-1 && gotYearDay <= wantYearDay) {
					t.Errorf("GetCandlesFor1NthDayBeforeNow() got = %v, want [%v; %v]", gotYearDay, wantYearDay-1, wantYearDay)
					t.FailNow()
				}
			}
		})
	}
}

func TestTradeEnv_GetAtLeastNLastCandles(t *testing.T) {
	e := New(utils.GetSandboxToken(), true)
	type args struct {
		figi           string
		candleInterval investapi.CandleInterval
		n              int
	}
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
		{
			name: "test2",
			args: args{
				figi:           "BBG006L8G4H1",
				candleInterval: investapi.CandleInterval_CANDLE_INTERVAL_5_MIN,
				n:              1,
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				figi:           "",
				candleInterval: investapi.CandleInterval_CANDLE_INTERVAL_1_MIN,
				n:              1,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.GetAtLeastNLastCandles(tt.args.figi, tt.args.candleInterval, tt.args.n)
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
