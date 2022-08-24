package utils

import (
	"reflect"
	"testing"
	"tinkoff-invest-contest/internal/client/investapi"
)

func TestFloatToQuotation(t *testing.T) {
	type args struct {
		value float64
	}
	tests := []struct {
		name string
		args args
		want *investapi.Quotation
	}{
		{
			name: "test1",
			args: args{
				value: 123.45,
			},
			want: &investapi.Quotation{
				Units: 123,
				Nano:  450000000,
			},
		},
		{
			name: "test2",
			args: args{
				value: 123.04,
			},
			want: &investapi.Quotation{
				Units: 123,
				Nano:  40000000,
			},
		},
		{
			name: "test3",
			args: args{
				value: -100.001,
			},
			want: &investapi.Quotation{
				Units: -100,
				Nano:  -1000000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatToQuotation(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FloatToQuotation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuotationToFloat(t *testing.T) {
	type args struct {
		q *investapi.Quotation
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test1",
			args: args{
				q: &investapi.Quotation{
					Units: 123,
					Nano:  450000000,
				},
			},
			want: 123.45,
		},
		{
			name: "test2",
			args: args{
				q: &investapi.Quotation{
					Units: 123,
					Nano:  40000000,
				},
			},
			want: 123.04,
		},
		{
			name: "test3",
			args: args{
				q: &investapi.Quotation{
					Units: -100,
					Nano:  -1000000,
				},
			},
			want: -100.001,
		},
		{
			name: "test4",
			args: args{
				q: &investapi.Quotation{
					Units: 0,
					Nano:  10,
				},
			},
			want: 0.00000001,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QuotationToFloat(tt.args.q); got != tt.want {
				t.Errorf("QuotationToFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoundQuotation(t *testing.T) {
	type args struct {
		q                 *investapi.Quotation
		minPriceIncrement *investapi.Quotation
	}
	tests := []struct {
		name string
		args args
		want *investapi.Quotation
	}{
		{
			name: "test1",
			args: args{
				q: &investapi.Quotation{
					Units: 123,
					Nano:  456000000,
				},
				minPriceIncrement: &investapi.Quotation{
					Units: 0,
					Nano:  10000000,
				},
			},
			want: &investapi.Quotation{
				Units: 123,
				Nano:  460000000,
			},
		},
		{
			name: "test2",
			args: args{
				q: &investapi.Quotation{
					Units: 123,
					Nano:  456000000,
				},
				minPriceIncrement: &investapi.Quotation{
					Units: 0,
					Nano:  1000000,
				},
			},
			want: &investapi.Quotation{
				Units: 123,
				Nano:  456000000,
			},
		},
		{
			name: "test3",
			args: args{
				q: &investapi.Quotation{
					Units: 0,
					Nano:  -123456789,
				},
				minPriceIncrement: &investapi.Quotation{
					Units: 0,
					Nano:  10,
				},
			},
			want: &investapi.Quotation{
				Units: 0,
				Nano:  -123456790,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoundQuotation(tt.args.q, tt.args.minPriceIncrement); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoundQuotation() = %v, want %v", got, tt.want)
			}
		})
	}
}
