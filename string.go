// utility/strings.go
package Utility

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"reflect"
	"sort"
	"unicode"

	"github.com/kalafut/imohash"
	"github.com/pborman/uuid"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Contains checks if a slice contains a given string.
func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

// Remove removes an element from a slice by index.
func Remove(s []string, index int) ([]string, error) {
	if index >= len(s) {
		return nil, errors.New("out of range error")
	}
	return append(s[:index], s[index+1:]...), nil
}

// RemoveString removes the first occurrence of r from slice s.
func RemoveString(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// InsertStringAt inserts str at position pos in slice arr.
func InsertStringAt(pos int, str string, arr *[]string) {
	*arr = append(*arr, "")
	for i := len(*arr) - 1; i > pos; i-- {
		(*arr)[i] = (*arr)[i-1]
	}
	(*arr)[pos] = str
}

// InsertIntAt inserts an int at a given position in slice arr.
func InsertIntAt(pos int, val int, arr *[]int) {
	*arr = append(*arr, 0)
	for i := len(*arr) - 1; i > pos; i-- {
		(*arr)[i] = (*arr)[i-1]
	}
	(*arr)[pos] = val
}

// InsertInt64At inserts an int64 at a given position in slice arr.
func InsertInt64At(pos int, val int64, arr *[]int64) {
	*arr = append(*arr, 0)
	for i := len(*arr) - 1; i > pos; i-- {
		(*arr)[i] = (*arr)[i-1]
	}
	(*arr)[pos] = val
}

// InsertBoolAt inserts a bool at a given position in slice arr.
func InsertBoolAt(pos int, val bool, arr *[]bool) {
	*arr = append(*arr, false)
	for i := len(*arr) - 1; i > pos; i-- {
		(*arr)[i] = (*arr)[i-1]
	}
	(*arr)[pos] = val
}

// RemoveAccent strips accents/diacritics from text.
func RemoveAccent(text string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ := transform.String(t, text)
	return s
}

// SortStrings returns a new sorted copy of the input slice.
func SortStrings(s []string) []string {
	result := make([]string, len(s))
	copy(result, s)
	sort.Strings(result)
	return result
}


// Create a random uuid value.
func RandomUUID() string {
	return uuid.NewRandom().String()
}

// Create a MD5 hash value with UUID format.
func GenerateUUID(val string) string {
	return uuid.NewMD5(uuid.NameSpace_DNS, []byte(val)).String()
}

/**
 * GetMD5Hash returns the MD5 hash of the input text.
 */
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

/**
 * Recursive function that return the checksum value.
 */
func GetChecksum(values interface{}) string {
	var checksum string

	if reflect.TypeOf(values).String() == "map[string]interface {}" {
		var keys []string
		for k, _ := range values.(map[string]interface{}) {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if values.(map[string]interface{})[key] != nil {
				checksum += GetChecksum(values.(map[string]interface{})[key])
			}
		}

	} else if reflect.TypeOf(values).String() == "[]interface {}" {

		for i := 0; i < len(values.([]interface{})); i++ {
			if values.([]interface{})[i] != nil {
				checksum += GetChecksum(values.([]interface{})[i])
			}
		}

	} else if reflect.TypeOf(values).String() == "[]map[string]interface {}" {
		for i := 0; i < len(values.([]map[string]interface{})); i++ {
			if values.([]map[string]interface{})[i] != nil {
				checksum += GetChecksum(values.([]map[string]interface{})[i])
			}
		}
	} else if reflect.TypeOf(values).String() == "[]string" {
		for i := 0; i < len(values.([]string)); i++ {
			checksum += GetChecksum(values.([]string)[i])
		}
	} else {
		// here the value must be a single value...
		checksum += ToString(values)
	}

	return GetMD5Hash(checksum)
}

const filechunk = 8192 // we settle for 8KB
func CreateFileChecksum(path string) string {
	checksum, _ := imohash.SumFile(path)
	return GetMD5Hash(string(checksum[:]))
}

func CreateDataChecksum(data []byte) string {
	checksum := imohash.Sum(data)
	return GetMD5Hash(string(checksum[:]))
}