package csvee

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewReader initializes and a new reader and verifies the resulting Reader is as expected
func TestNewReader(t *testing.T) {

	columnNames := []string{"1", "2", "3"}

	reader := NewReader(strings.NewReader("a,b,c"), columnNames)

	assert.NotNil(t, reader)
	assert.Exactly(t, columnNames, reader.ColumnNames)
}

type readTo struct {
	F float64
	I int
	B bool
	S string
}

// TestReader_Read reads from a Reader and verifies the resulting struct is as expected
func TestReader_Read(t *testing.T) {

	var testCases = []struct {
		name          string
		inData        string
		inColumnNames []string
		expData       readTo
		expErr        bool
		expErrText    string
	}{
		{
			name:          "success",
			inData:        `29.4,3,true,"hello ""you"""`,
			inColumnNames: []string{"F", "I", "B", "S"},
			expData:       readTo{F: 29.4, I: 3, B: true, S: `hello "you"`},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			reader := NewReader(strings.NewReader(tt.inData), tt.inColumnNames)
			var actualData readTo
			err := reader.Read(&actualData)

			// Hard stop if the expectation of an error isn't fulfilled.
			require.Equal(t, tt.expErr, err != nil, err)
			if err != nil {
				assert.Equal(t, tt.expErrText, err.Error())
				return
			}

			assert.Equal(t, tt.expData.F, actualData.F)
			assert.Equal(t, tt.expData.I, actualData.I)
			assert.Equal(t, tt.expData.B, actualData.B)
			assert.Equal(t, tt.expData.S, actualData.S)
		})
	}

}

// TestReader_ReadAll reads from a Reader and verifies the resulting struct slice is as expected
