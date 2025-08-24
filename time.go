// utility/time.go
package Utility

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MakeTimestamp returns the current Unix timestamp in seconds.
func MakeTimestamp() int64 {
	return time.Now().Unix()
}

// DateTimeFromString parses a date string with a given layout.
func DateTimeFromString(str string, layout string) (time.Time, error) {
	return time.Parse(layout, str)
}

// MatchISO8601_Time parses an ISO8601 time string into a time.Time (UTC).
func MatchISO8601_Time(str string) (*time.Time, error) {
	exp := regexp.MustCompile(ISO_8601_TIME_PATTERN)
	match := exp.FindStringSubmatch(str)
	if len(match) == 0 {
		return nil, errors.New(str + " not match iso 8601")
	}

	var hour, minute, second, miliSecond int
	for i, name := range exp.SubexpNames() {
		if i != 0 && match[i] != "" {
			switch name {
			case "hour":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				hour = int(val)
			case "minute":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				minute = int(val)
			case "second":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				second = int(val)
			case "ms":
				val, _ := strconv.ParseFloat(match[i], 64)
				miliSecond = int(val * 1000)
			}
		}
	}
	t := time.Date(0, time.Month(0), 0, hour, minute, second, miliSecond, time.UTC)
	return &t, nil
}

// MatchISO8601_Date parses an ISO8601 date string into a time.Time (UTC).
func MatchISO8601_Date(str string) (*time.Time, error) {
	exp := regexp.MustCompile(ISO_8601_DATE_PATTERN)
	match := exp.FindStringSubmatch(str)
	if len(match) == 0 {
		return nil, errors.New(str + " not match iso 8601")
	}

	var year, month, day int
	for i, name := range exp.SubexpNames() {
		if i != 0 && match[i] != "" {
			switch name {
			case "year":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				year = int(val)
			case "month":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				month = int(val)
			case "day":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				day = int(val)
			}
		}
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t, nil
}

// MatchISO8601_DateTime parses an ISO8601 datetime string into a time.Time (UTC).
func MatchISO8601_DateTime(str string) (*time.Time, error) {
	exp := regexp.MustCompile(ISO_8601_DATE_TIME_PATTERN)
	match := exp.FindStringSubmatch(str)
	if len(match) == 0 {
		return nil, errors.New(str + " not match iso 8601")
	}

	var year, month, day, hour, minute, second, miliSecond int
	for i, name := range exp.SubexpNames() {
		if i != 0 && match[i] != "" {
			switch name {
			case "year":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				year = int(val)
			case "month":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				month = int(val)
			case "day":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				day = int(val)
			case "hour":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				hour = int(val)
			case "minute":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				minute = int(val)
			case "second":
				val, _ := strconv.ParseInt(match[i], 10, 64)
				second = int(val)
			case "ms":
				val, _ := strconv.ParseFloat(match[i], 64)
				miliSecond = int(val * 1000)
			}
		}
	}
	t := time.Date(year, time.Month(month), day, hour, minute, second, miliSecond, time.UTC)
	return &t, nil
}

