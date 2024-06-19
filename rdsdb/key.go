package rdsdb

import (
	"fmt"
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
func ConcatedKeys(fields ...interface{}) string {
	var results strings.Builder
	//for each field ,it it's type if float64 or float32,but it's value is integer,then convert it to int
	for i, field := range fields {
		field_value := field
		if f64, ok := field.(float64); ok && f64 == float64(int64(f64)) {
			field_value = int64(f64)
		} else if f32, ok := field.(float32); ok && f32 == float32(int32(f32)) {
			field_value = int32(f32)
		}
		if i > 0 {
			results.WriteString(":")
		}
		results.WriteString(fmt.Sprintf("%v", field_value))
	}
	return results.String()
}
