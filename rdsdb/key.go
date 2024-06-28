package rdsdb

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func CatYearMonthDay(tm time.Time) string {
	//year is 4 digits, month is 2 digits, day is 2 digits
	return fmt.Sprintf("YMD_%04v%02v%02v", tm.Year(), int(tm.Month()), tm.Day())
}
func CatYearMonth(tm time.Time) string {
	//year is 4 digits, month is 2 digits
	return fmt.Sprintf("YM_%04v%02v", tm.Year(), int(tm.Month()))
}
func CatYear(tm time.Time) string {
	//year is 4 digits
	return fmt.Sprintf("Y_%04v", tm.Year())
}
func CatYearWeek(tm time.Time) string {
	tm = tm.UTC()
	isoYear, isoWeek := tm.ISOWeek()
	//year is 4 digits, week is 2 digits
	return fmt.Sprintf("YW_%04v%02v", isoYear, isoWeek)
}
func ConcatedKeys(key string, fields ...interface{}) string {
	var results strings.Builder
	results.WriteString(key)
	//for each field ,it it's type if float64 or float32,but it's value is integer,then convert it to int
	for _, field := range fields {
		//if field is nil,skip it
		if field == nil {
			continue
		}
		results.WriteString(":")
		if f64, ok := field.(float64); ok && f64 == float64(int64(f64)) {
			results.WriteString(strconv.FormatInt(int64(f64), 10))
		} else if f32, ok := field.(float32); ok && f32 == float32(int32(f32)) {
			results.WriteString(strconv.FormatInt(int64(f32), 10))
		} else if s, ok := field.(string); ok {
			results.WriteString(s)
		} else {
			results.WriteString(fmt.Sprintf("%v", field))
		}
	}
	return results.String()
}
