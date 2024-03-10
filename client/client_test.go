package client

import (
	"fmt"
	"testing"
)

func TestTrueRandFloat(t *testing.T) {
	type args struct {
		min float64
		max float64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "case1",
			args: args{
				min: 5,
				max: 15,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num := TrueRandFloat(tt.args.min, tt.args.max)
			fmt.Printf("TrueRandFloat() = %v; want %v\n", num, tt.args)
			if num < tt.args.min || num > tt.args.max {
				t.Errorf("TrueRandFloat() = %v; want %v", TrueRandFloat(tt.args.min, tt.args.max), tt.args)
			}
		})
	}
}
