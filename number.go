// utility/numbers.go
package Utility

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ToString converts many primitive/interface types into a string.
func ToString(value interface{}) string {
	var str string
	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		str = value.(string)
	case reflect.Int:
		str = strconv.Itoa(ToInt(value))
	case reflect.Int8:
		str = strconv.Itoa(int(value.(int8)))
	case reflect.Int16:
		str = strconv.Itoa(int(value.(int16)))
	case reflect.Int32:
		str = strconv.Itoa(int(value.(int32)))
	case reflect.Int64:
		str = strconv.Itoa(int(value.(int64)))
	case reflect.Uint8:
		str = strconv.Itoa(int(value.(uint8)))
	case reflect.Uint16:
		str = strconv.Itoa(int(value.(uint16)))
	case reflect.Uint32:
		str = strconv.Itoa(int(value.(uint32)))
	case reflect.Uint64:
		str = strconv.Itoa(int(value.(uint64)))
	case reflect.Float32:
		str = strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32)
	case reflect.Float64:
		str = strconv.FormatFloat(value.(float64), 'f', -1, 64)
	case reflect.Bool:
		str = strconv.FormatBool(value.(bool))
	default:
		t := reflect.TypeOf(value).String()
		if t == "[]uint8" {
			str = string(value.([]uint8))
		} else if t == "*errors.errorString" || t == "*errors.Error" {
			str = value.(error).Error()
		} else if t == "[]string" {
			for i, v := range value.([]string) {
				str += v
				if i < len(value.([]string))-1 {
					str += " "
				}
			}
		} else if t == "map[string]interface {}" {
			data, err := json.Marshal(value)
			if err == nil {
				return string(data)
			}
			return "{}"
		} else {
			log.Panicln("Value with type:", reflect.TypeOf(value).String(), "cannot be converted to string")
		}
	}
	return strings.TrimSpace(str)
}

// ToInt converts many primitive/interface types into int.
func ToInt(value interface{}) int {
	if value == nil {
		return 0
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		val, _ := strconv.Atoi(value.(string))
		return val
	case reflect.Int:
		return value.(int)
	case reflect.Int8:
		return int(value.(int8))
	case reflect.Int16:
		return int(value.(int16))
	case reflect.Int32:
		return int(value.(int32))
	case reflect.Int64:
		return int(value.(int64))
	case reflect.Float32:
		return int(value.(float32))
	case reflect.Float64:
		return int(value.(float64))
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	default:
		if reflect.TypeOf(value).String() == "[]uint8" {
			return int(binary.BigEndian.Uint64(value.([]uint8)))
		}
	}
	log.Panicln("Value with type:", reflect.TypeOf(value).String(), "cannot be converted to int")
	return 0
}

// IsBool checks if the value is or can be parsed as bool.
func IsBool(value interface{}) bool {
	if reflect.TypeOf(value).Kind() == reflect.Bool {
		return true
	} else if reflect.TypeOf(value).Kind() == reflect.String {
		_, err := strconv.ParseBool(value.(string))
		return err == nil
	}
	return false
}

// ToBool converts value to bool (false if parse fails).
func ToBool(value interface{}) bool {
	if reflect.TypeOf(value).Kind() == reflect.Bool {
		return value.(bool)
	} else if reflect.TypeOf(value).Kind() == reflect.String {
		b, err := strconv.ParseBool(value.(string))
		if err == nil {
			return b
		}
	}
	return false
}

// IsNumeric checks if value is a numeric type.
func IsNumeric(value interface{}) bool {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.String, reflect.Bool:
		return false
	default:
		if reflect.TypeOf(value).String() == "time.Time" {
			return true
		}
	}
	return false
}

// ToNumeric converts value into float64 (bool -> 0/1, time -> unix timestamp).
func ToNumeric(value interface{}) float64 {
	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		val, _ := strconv.ParseFloat(value.(string), 64)
		return val
	case reflect.Int:
		return float64(value.(int))
	case reflect.Int8:
		return float64(value.(int8))
	case reflect.Int16:
		return float64(value.(int16))
	case reflect.Int32:
		return float64(value.(int32))
	case reflect.Int64:
		return float64(value.(int64))
	case reflect.Float32:
		return float64(value.(float32))
	case reflect.Float64:
		return value.(float64)
	case reflect.Bool:
		if value.(bool) {
			return 1.0
		}
		return 0.0
	default:
		if reflect.TypeOf(value).String() == "time.Time" {
			return float64(value.(time.Time).Unix())
		}
	}
	log.Panicln("Value with type:", reflect.TypeOf(value).String(), "cannot be converted to float64")
	return 0
}

// Round rounds float64 to n decimals using bankers rounding.
func Round(x float64, n int) float64 {
	pow := math.Pow(10, float64(n))
	if math.Abs(x*pow) > 1e17 {
		return x
	}
	v, frac := math.Modf(x * pow)
	if x > 0.0 {
		if frac > 0.5 || (frac == 0.5 && uint64(v)%2 != 0) {
			v += 1.0
		}
	} else {
		if frac < -0.5 || (frac == -0.5 && uint64(v)%2 != 0) {
			v -= 1.0
		}
	}
	return v / pow
}

// Less compares two values of the same type and reports val0 < val1.
func Less(val0, val1 interface{}) bool {
	if val0 == nil || val1 == nil {
		return true
	}
	switch reflect.TypeOf(val0).Kind() {
	case reflect.String:
		return val0.(string) < val1.(string)
	case reflect.Int:
		return val0.(int) < val1.(int)
	case reflect.Int8:
		return val0.(int8) < val1.(int8)
	case reflect.Int16:
		return val0.(int16) < val1.(int16)
	case reflect.Int32:
		return val0.(int32) < val1.(int32)
	case reflect.Int64:
		return val0.(int64) < val1.(int64)
	case reflect.Float32:
		return val0.(float32) < val1.(float32)
	case reflect.Float64:
		return val0.(float64) < val1.(float64)
	default:
		log.Println("Value with type:", reflect.TypeOf(val0).String(), "cannot be compared")
	}
	return false
}

