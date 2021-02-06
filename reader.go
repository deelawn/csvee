package csvee

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"reflect"
	"strings"
)

// Reader embeds *csv.Reader and contains the column names of the CSV data that is to be read.
type Reader struct {
	CSVReader   *csv.Reader
	ColumnNames []string
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader, columnNames []string) *Reader {

	columnNamesCopy := make([]string, len(columnNames))
	_ = copy(columnNamesCopy, columnNames)

	return &Reader{
		CSVReader:   csv.NewReader(r),
		ColumnNames: columnNamesCopy,
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
		}

		// If it is a slice then assign the json array representation to fieldValue
		if fieldSliceType != nil {

			fieldValue = "["

			// Quote each value in the slice in the case of strings
			if fieldSliceType.Kind() == reflect.String {
				sliceValues := strings.Split(field, ",")
				for i := 0; i < len(sliceValues); i++ {
					sliceValues[i] = `"` + sliceValues[i] + `"`
				}
				fieldValue += strings.Join(sliceValues, ",")
			} else {
				fieldValue = field
			}

			fieldValue += "]"
		}

		labeledFields[i] = `"` + r.ColumnNames[i] + `":` + fieldValue
	}

	// Build the JSON
	jsonRecord := []byte("{" + strings.Join(labeledFields, ",") + "}")

	// Try to Unmarshal it to the provided interface
	return json.Unmarshal(jsonRecord, v)
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

	isValidType = typeIsValid(t)
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
