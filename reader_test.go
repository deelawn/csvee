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
	F  float64
	I  int
	B  bool
	S  string
	IP *int
	IA []int
	SA []string
}

// TestReader_Read reads from a Reader and verifies the resulting struct is as expected
func TestReader_Read(t *testing.T) {

	var intPtr *int = new(int)
	*intPtr = 9

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
			inData:        `29.4,3,true,"hello ""you""",9,"8,4,3,5","this,is,not,a,test"`,
			inColumnNames: []string{"F", "I", "B", "S", "IP", "IA", "SA"},
			expData: readTo{
				F:  29.4,
				I:  3,
				B:  true,
				S:  `hello "you"`,
				IP: intPtr,
				IA: []int{8, 4, 3, 5},
				SA: []string{"this", "is", "not", "a", "test"},
			},
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

			if tt.expData.IP != nil {
				assert.NotNil(t, actualData.IP)
				if actualData.IP != nil {
					assert.Equal(t, *tt.expData.IP, *actualData.IP)
				}
			} else {
				assert.Nil(t, actualData.IP)
			}

			assert.Exactly(t, tt.expData.IA, actualData.IA)
			assert.Exactly(t, tt.expData.SA, actualData.SA)
		})
	}

}

// TestReader_ReadAll reads from a Reader and verifies the resulting struct slice is as expected
