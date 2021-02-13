package csvee

import "errors"

const TimeFormatUnix string = "unix"

var (
	ErrColumnNamesMismatch   = errors.New("The number of column names does not match the number of fieldsin the record.")
	ErrUnsupportedTargetType = errors.New("Target interface must be of type struct or map.")
	ErrInvalidFieldType      = errors.New("Struct field type must be int*, float*, bool, string, time, or a slice.")
)
