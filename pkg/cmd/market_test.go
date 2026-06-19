package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestRootCommandContainsMarket(t *testing.T) {
	var found bool
	for _, child := range Command.Commands {
		if child.Name == "market" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestMarketCommandContainsTick(t *testing.T) {
	found := map[string]bool{}
	for _, child := range marketCmd.Commands {
		found[child.Name] = true
	}
	assert.True(t, found["tick"])
}

func TestMarketTickCommandContainsFlags(t *testing.T) {
	found := map[string]bool{}
	for _, flag := range marketTickCmd.Flags {
		for _, name := range flag.Names() {
			found[name] = true
		}
	}
	assert.True(t, found["price"])
}

func TestBuildMarketTickOutputBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		price       int64
		tickSize    int64
		lowerPrice  int64
		upperPrice  int64
		isValidTick bool
	}{
		{name: "one", price: 1, tickSize: 1, lowerPrice: 1, upperPrice: 1, isValidTick: true},
		{name: "under 2000", price: 1999, tickSize: 1, lowerPrice: 1999, upperPrice: 1999, isValidTick: true},
		{name: "2000", price: 2000, tickSize: 5, lowerPrice: 2000, upperPrice: 2000, isValidTick: true},
		{name: "2001", price: 2001, tickSize: 5, lowerPrice: 2000, upperPrice: 2005, isValidTick: false},
		{name: "4999", price: 4999, tickSize: 5, lowerPrice: 4995, upperPrice: 5000, isValidTick: false},
		{name: "5000", price: 5000, tickSize: 10, lowerPrice: 5000, upperPrice: 5000, isValidTick: true},
		{name: "19999", price: 19999, tickSize: 10, lowerPrice: 19990, upperPrice: 20000, isValidTick: false},
		{name: "20000", price: 20000, tickSize: 50, lowerPrice: 20000, upperPrice: 20000, isValidTick: true},
		{name: "49999", price: 49999, tickSize: 50, lowerPrice: 49950, upperPrice: 50000, isValidTick: false},
		{name: "50000", price: 50000, tickSize: 100, lowerPrice: 50000, upperPrice: 50000, isValidTick: true},
		{name: "199999", price: 199999, tickSize: 100, lowerPrice: 199900, upperPrice: 200000, isValidTick: false},
		{name: "200000", price: 200000, tickSize: 500, lowerPrice: 200000, upperPrice: 200000, isValidTick: true},
		{name: "499999", price: 499999, tickSize: 500, lowerPrice: 499500, upperPrice: 500000, isValidTick: false},
		{name: "500000", price: 500000, tickSize: 1000, lowerPrice: 500000, upperPrice: 500000, isValidTick: true},
		{name: "sample", price: 353333, tickSize: 500, lowerPrice: 353000, upperPrice: 353500, isValidTick: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := buildMarketTickOutput(tt.price)
			assert.Equal(t, tt.price, output.Price)
			assert.Equal(t, tt.tickSize, output.TickSize)
			assert.Equal(t, tt.lowerPrice, output.LowerPrice)
			assert.Equal(t, tt.upperPrice, output.UpperPrice)
			assert.Equal(t, tt.isValidTick, output.IsValidTick)
		})
	}
}

func TestMarketTickOutputUsesSchema(t *testing.T) {
	output := buildMarketTickOutput(353333)

	data, err := json.Marshal(output)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, map[string]struct{}{
		"price":         {},
		"tick_size":     {},
		"lower_price":   {},
		"upper_price":   {},
		"is_valid_tick": {},
	}, keySet(decoded))
	assert.Equal(t, float64(353333), decoded["price"])
	assert.Equal(t, float64(500), decoded["tick_size"])
	assert.Equal(t, float64(353000), decoded["lower_price"])
	assert.Equal(t, float64(353500), decoded["upper_price"])
	assert.Equal(t, false, decoded["is_valid_tick"])
}

func TestParseMarketTickPriceRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name  string
		price string
		want  string
	}{
		{name: "missing", price: "", want: "missing price"},
		{name: "zero", price: "0", want: "invalid price"},
		{name: "negative", price: "-1", want: "invalid price"},
		{name: "decimal", price: "353333.5", want: "invalid price"},
		{name: "non numeric", price: "nope", want: "invalid price"},
		{name: "too large", price: "9223372036854775807", want: "no greater than 9223372036854774808"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseMarketTickPrice(tt.price)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestMarketTickValidationRejectsBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := runMarketTick(context.Background(), marketTickOptions{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing price")
	assert.NotContains(t, err.Error(), "stock config set")
}

func TestMarketTickRejectsUnexpectedArgs(t *testing.T) {
	cmd := &cli.Command{}
	cmd.Commands = []*cli.Command{&marketCmd}
	err := cmd.Run(context.Background(), []string{"stock", "market", "tick", "--price", "353333", "extra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected extra arguments")
}
