package retry

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetryCallFunc(t *testing.T) {
	type args struct {
		Retry       int
		RetryPeriod int
		Timeout     int
	}
	tests := []struct {
		name string
		arg  args
		fn   func(param interface{})
	}{
		// TODO: Add test cases.
		{
			name: "RetryCallFunc",
			arg: args{
				Retry:       10,
				RetryPeriod: 1000,
				Timeout:     20000,
			},
			fn: func(param interface{}) {
				p := param.(args)
				_, err := RetryCallFunc(context.TODO(), func(ctx context.Context, param interface{}) (interface{}, error) {
					fmt.Printf("retry\n")
					return nil, fmt.Errorf("test")
				}, param, p.Retry, p.RetryPeriod, p.Timeout)
				fmt.Printf("err: %v", err)
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(tt.arg)
		})
	}
}
