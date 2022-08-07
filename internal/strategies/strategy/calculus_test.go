package strategy

import "testing"

func TestIsAroundPoint(t *testing.T) {
	type args struct {
		samplePoint float64
		refPoint    float64
		deviation   float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				samplePoint: 89,
				refPoint:    100,
				deviation:   0.1,
			},
			want: false,
		},
		{
			name: "test2",
			args: args{
				samplePoint: 1001.01,
				refPoint:    1000,
				deviation:   0.001,
			},
			want: false,
		},
		{
			name: "test3",
			args: args{
				samplePoint: 1001.01,
				refPoint:    1000,
				deviation:   0.002,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAroundPoint(tt.args.samplePoint, tt.args.refPoint, tt.args.deviation); got != tt.want {
				t.Errorf("IsAroundPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBetweenIncl(t *testing.T) {
	type args struct {
		samplePoint float64
		bound1      float64
		bound2      float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				samplePoint: 10,
				bound1:      9,
				bound2:      11,
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				samplePoint: 8.999999,
				bound1:      9,
				bound2:      9.000001,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBetweenIncl(tt.args.samplePoint, tt.args.bound1, tt.args.bound2); got != tt.want {
				t.Errorf("IsBetweenIncl() = %v, want %v", got, tt.want)
			}
		})
	}
}
