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

	reader, err := NewReader(strings.NewReader("a,b,c"), &ReaderOptions{ColumnNames: columnNames})

	// Test without column formats
	require.NotNil(t, reader)
	require.NoError(t, err)
	assert.Exactly(t, columnNames, reader.ColumnNames)

	reader, err = NewReader(
		strings.NewReader("a,b,c"),
		&ReaderOptions{
			ColumnNames:   columnNames,
			ColumnFormats: map[string]string{"1": "a", "2": "b"},
		},
	)

	// Test with column formats
	require.NotNil(t, reader)
	require.NoError(t, err)
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
		inReadHeaders   bool
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
		{
			name: "success reading headers",
			inData: "F,I,B,S,IP,IA,SA,Tu,T\n" +
				`29.4,3,true,"hello ""you""",9,"8,4,3,5","this,is,not,a,test",1613235342,1991-04-05T11:11:11Z`,
			inColumnFormats: map[string]string{"Tu": TimeFormatUnix},
			inReadHeaders:   true,
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

			reader, err := NewReader(
				strings.NewReader(tt.inData),
				&ReaderOptions{
					ColumnNames:   tt.inColumnNames,
					ColumnFormats: tt.inColumnFormats,
					ReadHeaders:   tt.inReadHeaders,
				},
			)

			require.NoError(t, err)

			var actualData readTo
			err = reader.Read(&actualData)

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
			name: "success",
			// + here is used purely for formatting. See below: reader.CSVReader.Comment = '+'
			inData: `+
				29.4,3,true,"hello ""you""",9,"8,4,3,5","this,is,not,a,test",1613235342,1991-04-05T11:11:11Z
				999.12,4,false,lorem ipsum...,9,"-9,3","this,might,be,a,test",1513235342,2007-05-27T15:00:00-05:00`,
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
				{
					F:  999.12,
					I:  4,
					B:  false,
					S:  "lorem ipsum...",
					IP: intPtr,
					IA: []int{-9, 3},
					SA: []string{"this", "might", "be", "a", "test"},
					Tu: time.Unix(1513235342, 0),
					T:  time.Date(2007, time.May, 27, 15, 0, 0, 0, time.FixedZone("America/New_York", -60*60*5)),
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			reader, err := NewReader(
				strings.NewReader(tt.inData),
				&ReaderOptions{
					ColumnNames:   tt.inColumnNames,
					ColumnFormats: tt.inColumnFormats,
				},
			)

			require.NoError(t, err)

			reader.CSVReader.TrimLeadingSpace = true
			reader.CSVReader.Comment = '+'
			actualData := []readTo{}
			err = reader.ReadAll(&actualData)

			// Hard stop if the expectation of an error isn't fulfilled.
			require.Equal(t, tt.expErr, err != nil, err)
			if err != nil {
				assert.Equal(t, tt.expErrText, err.Error())
				return
			}

			require.Len(t, actualData, 2)

			makeAssertions := func(idx int) {
				assert.Equal(t, tt.expData[idx].F, actualData[idx].F)
				assert.Equal(t, tt.expData[idx].I, actualData[idx].I)
				assert.Equal(t, tt.expData[idx].B, actualData[idx].B)
				assert.Equal(t, tt.expData[idx].S, actualData[idx].S)

				if tt.expData[idx].IP != nil {
					assert.NotNil(t, actualData[idx].IP)
					if actualData[idx].IP != nil {
						assert.Equal(t, *tt.expData[idx].IP, *actualData[idx].IP)
					}
				} else {
					assert.Nil(t, actualData[idx].IP)
				}

				assert.Exactly(t, tt.expData[idx].IA, actualData[idx].IA)
				assert.Exactly(t, tt.expData[idx].SA, actualData[idx].SA)
				assert.Equal(t, tt.expData[idx].Tu.Unix(), actualData[idx].Tu.Unix())
				assert.Equal(t, tt.expData[idx].T.Unix(), actualData[idx].T.Unix())
			}

			makeAssertions(0)
			makeAssertions(1)
		})
	}

}
