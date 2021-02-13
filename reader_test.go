package csvee

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewReader initializes and a new reader and verifies the resulting Reader is as expected
func TestNewReader(t *testing.T) {

	columnNames := []string{"1", "2", "3"}

	reader := NewReader(strings.NewReader("a,b,c"), columnNames)

	// Test without column formats
	require.NotNil(t, reader)
	assert.Exactly(t, columnNames, reader.ColumnNames)

	reader = NewReader(strings.NewReader("a,b,c"), columnNames, map[string]string{"1": "a", "2": "b"})

	// Test with column formats
	require.NotNil(t, reader)
	assert.Exactly(t, columnNames, reader.ColumnNames)
	require.NotNil(t, reader.ColumnFormats)
	assert.Equal(t, "a", reader.ColumnFormats["1"])
	assert.Equal(t, "b", reader.ColumnFormats["2"])
}

type readTo struct {
	F  float64
	I  int
	B  bool
	S  string
	IP *int
	IA []int
	SA []string
	Tu time.Time
	T  time.Time
}

// TestReader_Read reads from a Reader and verifies the resulting struct is as expected
func TestReader_Read(t *testing.T) {

	var intPtr *int = new(int)
	*intPtr = 9

	var testCases = []struct {
		name            string
		inData          string
		inColumnNames   []string
		inColumnFormats map[string]string
		expData         readTo
		expErr          bool
		expErrText      string
	}{
		{
			name:            "success",
			inData:          `29.4,3,true,"hello ""you""",9,"8,4,3,5","this,is,not,a,test",1613235342,1991-04-05T11:11:11Z`,
			inColumnNames:   []string{"F", "I", "B", "S", "IP", "IA", "SA", "Tu", "T"},
			inColumnFormats: map[string]string{"Tu": TimeFormatUnix},
			expData: readTo{
				F:  29.4,
				I:  3,
				B:  true,
				S:  `hello "you"`,
				IP: intPtr,
				IA: []int{8, 4, 3, 5},
				SA: []string{"this", "is", "not", "a", "test"},
				Tu: time.Unix(1613235342, 0),
				T:  time.Date(1991, time.April, 5, 11, 11, 11, 0, time.UTC),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			reader := NewReader(strings.NewReader(tt.inData), tt.inColumnNames, tt.inColumnFormats)
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
			assert.Equal(t, tt.expData.Tu.Unix(), actualData.Tu.Unix())
			assert.Equal(t, tt.expData.T.Unix(), actualData.T.Unix())
		})
	}

}

// TestReader_ReadAll reads from a Reader and verifies the resulting struct slice is as expected
func TestReader_ReadAll(t *testing.T) {

	var intPtr *int = new(int)
	*intPtr = 9

	var testCases = []struct {
		name            string
		inData          string
		inColumnNames   []string
		inColumnFormats map[string]string
		expData         []readTo
		expErr          bool
		expErrText      string
	}{
		{
			name:            "success",
			inData:          `29.4,3,true,"hello ""you""",9,"8,4,3,5","this,is,not,a,test",1613235342,1991-04-05T11:11:11Z`,
			inColumnNames:   []string{"F", "I", "B", "S", "IP", "IA", "SA", "Tu", "T"},
			inColumnFormats: map[string]string{"Tu": TimeFormatUnix},
			expData: []readTo{
				{
					F:  29.4,
					I:  3,
					B:  true,
					S:  `hello "you"`,
					IP: intPtr,
					IA: []int{8, 4, 3, 5},
					SA: []string{"this", "is", "not", "a", "test"},
					Tu: time.Unix(1613235342, 0),
					T:  time.Date(1991, time.April, 5, 11, 11, 11, 0, time.UTC),
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			reader := NewReader(strings.NewReader(tt.inData), tt.inColumnNames, tt.inColumnFormats)
			actualData := []readTo{}
			err := reader.ReadAll(&actualData)

			// Hard stop if the expectation of an error isn't fulfilled.
			require.Equal(t, tt.expErr, err != nil, err)
			if err != nil {
				assert.Equal(t, tt.expErrText, err.Error())
				return
			}

			require.Len(t, actualData, 1)

			assert.Equal(t, tt.expData[0].F, actualData[0].F)
			assert.Equal(t, tt.expData[0].I, actualData[0].I)
			assert.Equal(t, tt.expData[0].B, actualData[0].B)
			assert.Equal(t, tt.expData[0].S, actualData[0].S)

			if tt.expData[0].IP != nil {
				assert.NotNil(t, actualData[0].IP)
				if actualData[0].IP != nil {
					assert.Equal(t, *tt.expData[0].IP, *actualData[0].IP)
				}
			} else {
				assert.Nil(t, actualData[0].IP)
			}

			assert.Exactly(t, tt.expData[0].IA, actualData[0].IA)
			assert.Exactly(t, tt.expData[0].SA, actualData[0].SA)
			assert.Equal(t, tt.expData[0].Tu.Unix(), actualData[0].Tu.Unix())
			assert.Equal(t, tt.expData[0].T.Unix(), actualData[0].T.Unix())
		})
	}

}
