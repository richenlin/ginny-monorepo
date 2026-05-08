package orm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseUrl(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		// TODO: Add test cases.
		{
			name: "1",
			dsn:  "mysql://127.0.0.1:3306/test?username=root&password=123456&charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
			u, err := c.parseUrl(tt.dsn)
			assert.NoError(t, err)
			fmt.Printf("%s\n", u)
		})
	}
}
