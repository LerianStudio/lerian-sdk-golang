package fees

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculationModelJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := CalculationModel{
		ApplicationRule: "maxBetweenTypes",
		Calculations: []Calculation{
			{Type: "flat", Value: "10.00"},
			{Type: "percentage", Value: "3.0"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CalculationModel

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ApplicationRule, decoded.ApplicationRule)
	require.Len(t, decoded.Calculations, 2)
	assert.Equal(t, "flat", decoded.Calculations[0].Type)
	assert.Equal(t, "10.00", decoded.Calculations[0].Value)
	assert.Equal(t, "percentage", decoded.Calculations[1].Type)
	assert.Equal(t, "3.0", decoded.Calculations[1].Value)
}

func TestCalculationJSONRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		calc Calculation
	}{
		{"flat fee", Calculation{Type: "flat", Value: "250.00"}},
		{"percentage", Calculation{Type: "percentage", Value: "2.5"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.calc)
			require.NoError(t, err)

			var decoded Calculation

			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.calc.Type, decoded.Type)
			assert.Equal(t, tt.calc.Value, decoded.Value)
		})
	}
}
