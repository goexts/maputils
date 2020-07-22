package extmap

import "bytes"

// Mapper ...
type Mapper interface {
	ToMap() Map
}

// XMLer ...
type XMLer interface {
	ParseXML([]byte) error
	ToXML() []byte
}

// JSONer ...
type JSONer interface {
	ParseJson([]byte) error
	ToJSON() (v []byte, err error)
}

// Stringer ...
type Stringer interface {
	String() string
}

/*ParseNumber parse interface to number */
func ParseNumber(v interface{}) (float64, bool) {
	switch v0 := v.(type) {
	case float64:
		return v0, true
	case float32:
		return float64(v0), true
	}
	return 0, false
}

/*ParseInt parse interface to int64 */
func ParseInt(v interface{}) (int64, bool) {
	switch v0 := v.(type) {
	case int:
		return int64(v0), true
	case int32:
		return int64(v0), true
	case int64:
		return v0, true
	case uint:
		return int64(v0), true
	case uint32:
		return int64(v0), true
	case uint64:
		return int64(v0), true
	case float64:
		return int64(v0), true
	case float32:
		return int64(v0), true
	default:
	}
	return 0, false
}

/*ParseString parse interface to string */
func ParseString(v interface{}) (string, bool) {
	switch v0 := v.(type) {
	case string:
		return v0, true
	case []byte:
		return string(v0), true
	case bytes.Buffer:
		return v0.String(), true
	default:
	}
	return "", false
}
