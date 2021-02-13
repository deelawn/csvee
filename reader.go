package csvee

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Reader embeds *csv.Reader and contains the column names of the CSV data that is to be read.
type Reader struct {
	CSVReader     *csv.Reader
	ColumnNames   []string
	ColumnFormats map[string]string
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader, columnNames []string, columnFormats ...map[string]string) *Reader {

	columnNamesCopy := make([]string, len(columnNames))
	_ = copy(columnNamesCopy, columnNames)

	lvColumnFormats := make(map[string]string)
	if columnFormats != nil {
		// Make a copy of whatever is passed in.
		for k, v := range columnFormats[0] {
			lvColumnFormats[k] = v
		}
	}

	return &Reader{
		CSVReader:     csv.NewReader(r),
		ColumnNames:   columnNamesCopy,
		ColumnFormats: lvColumnFormats,
	}
}

// Read reads the next line of the CSV and puts in into a struct.
func (r *Reader) Read(v interface{}) error {

	// The easiest way to convert a CSV line to a struct is to label the fields and utilize the
	// parser in encoding/json.

	// This handles any CSV read errors we might encounter.
	record, err := r.CSVReader.Read()
	if err != nil {
		return err
	}

	// It is possible to define behavior so that it processes as many fields as possible until one
	// of the two slices reaches its limit, but it isn't clear how that might work.
	if len(record) != len(r.ColumnNames) {
		return ErrColumnNamesMismatch
	}

	// v's type needs to be a struct or a map
	vType := getBaseType(reflect.TypeOf(v))
	if vType.Kind() != reflect.Struct && vType.Kind() != reflect.Map {
		return ErrUnsupportedTargetType
	}

	labeledFields := make([]string, len(record))
	for i, field := range record {

		// Get the struct field; skip this field if it doesn't exist in the struct.
		structField, exists := vType.FieldByName(r.ColumnNames[i])
		if !exists {
			continue
		}

		fieldType, fieldSliceType, isValidType := getFieldTypeInfo(structField.Type)
		if !isValidType {
			return ErrInvalidFieldType
		}

		fieldValue := field

		if fieldType.Kind() == reflect.String {
			fieldValue = strings.ReplaceAll(field, `"`, `\"`)
			fieldValue = `"` + fieldValue + `"`
		} else if isTimeType(fieldType) {
			if fieldValue, err = r.parseTime(field, i); err != nil {
				return err
			}
			fieldValue = `"` + fieldValue + `"`
		}

		// If it is a slice then assign the json array representation to fieldValue
		if fieldSliceType != nil {
			if fieldValue, err = r.buildSliceFieldValue(fieldSliceType, field, i); err != nil {
				return err
			}
		}

		labeledFields[i] = `"` + r.ColumnNames[i] + `":` + fieldValue
	}

	// Build the JSON
	jsonRecord := []byte("{" + strings.Join(labeledFields, ",") + "}")

	// Try to Unmarshal it to the provided interface
	return json.Unmarshal(jsonRecord, v)
}

func (r *Reader) parseTime(field string, column int) (string, error) {

	// First check whether a format was defined this time column
	format, exists := r.ColumnFormats[r.ColumnNames[column]]
	if !exists {
		// If no format exists, assume the string is formatted correctly as the default RFC3339 format
		return field, nil
	}

	var tm time.Time

	// Parse out income time strings from unix or other formats to time.Time
	if format == TimeFormatUnix {

		intField, err := strconv.ParseInt(field, 10, 0)
		if err != nil {
			return "", err
		}

		tm = time.Unix(intField, 0)

	} else {

		var err error
		if tm, err = time.Parse(format, field); err != nil {
			return "", err
		}
	}

	// Output times in RFC3339 format
	return tm.Format(time.RFC3339), nil
}

func (r *Reader) buildSliceFieldValue(t reflect.Type, field string, column int) (string, error) {

	fieldValue := "["

	if t.Kind() == reflect.String {
		sliceValues := strings.Split(field, ",")
		for i := 0; i < len(sliceValues); i++ {
			sliceValues[i] = `"` + sliceValues[i] + `"`
		}
		fieldValue += strings.Join(sliceValues, ",")
	} else if isTimeType(t) {
		sliceValues := strings.Split(field, ",")
		for i := 0; i < len(sliceValues); i++ {
			value, err := r.parseTime(sliceValues[i], column)
			if err != nil {
				return "", err
			}
			sliceValues[i] = `"` + value + `"`
		}
		fieldValue += strings.Join(sliceValues, ",")
	} else {
		fieldValue += field
	}

	fieldValue += "]"

	return fieldValue, nil
}

func getBaseType(t reflect.Type) reflect.Type {

	tp := t
	for {
		if tp.Kind() == reflect.Ptr {
			tp = tp.Elem()
			continue
		}
		break
	}

	return tp
}

func getFieldTypeInfo(t reflect.Type) (fieldType, sliceType reflect.Type, isValidType bool) {

	fieldType = getBaseType(t)
	if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
		sliceType = getBaseType(fieldType.Elem())
		isValidType = typeIsValid(sliceType)
		return
	}

	isValidType = typeIsValid(fieldType)
	return
}

func typeIsValid(t reflect.Type) bool {

	k := t.Kind()
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 ||
		k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 ||
		k == reflect.Float32 || k == reflect.Float64 || k == reflect.Bool || k == reflect.String || isTimeType(t)
}

func isTimeType(t reflect.Type) bool {

	return t.PkgPath() == "time" && t.Name() == "Time"
}
