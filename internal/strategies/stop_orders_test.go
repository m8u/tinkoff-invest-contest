package strategies

import (
	"testing"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

func TestTradeSignalStopOrder_IsTriggered(t *testing.T) {
	type fields struct {
		Direction    investapi.OrderDirection
		Type         investapi.StopOrderType
		TriggerPrice *investapi.Quotation
		LimitPrice   *investapi.Quotation
	}
	type args struct {
		price *investapi.Quotation
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test1",
			fields: fields{
				Direction:    investapi.OrderDirection_ORDER_DIRECTION_SELL,
				Type:         investapi.StopOrderType_STOP_ORDER_TYPE_TAKE_PROFIT,
				TriggerPrice: utils.FloatToQuotation(100.0),
			},
			args: args{
				price: utils.FloatToQuotation(101.0),
			},
			want: true,
		},
		{
			name: "test2",
			fields: fields{
				Direction:    investapi.OrderDirection_ORDER_DIRECTION_SELL,
				Type:         investapi.StopOrderType_STOP_ORDER_TYPE_TAKE_PROFIT,
				TriggerPrice: utils.FloatToQuotation(100.0),
			},
			args: args{
				price: utils.FloatToQuotation(99.99),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopOrder := &TradeSignalStopOrder{
				Direction:    tt.fields.Direction,
				Type:         tt.fields.Type,
				TriggerPrice: tt.fields.TriggerPrice,
				LimitPrice:   tt.fields.LimitPrice,
			}
			if got := stopOrder.IsTriggered(tt.args.price); got != tt.want {
				t.Errorf("IsTriggered() = %v, want %v", got, tt.want)
			}
		})
	}
}
