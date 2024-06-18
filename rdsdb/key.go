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
	results := make([]string, 0, len(fields)+1)

	//for each field ,it it's type if float64 or float32,but it's value is integer,then convert it to int
	for i, field := range fields {
		if f64, ok := field.(float64); ok && f64 == float64(int64(f64)) {
			results = append(results, fmt.Sprintf("%v", int64(field.(float64))))
		} else if f32, ok := field.(float32); ok && f32 == float32(int32(f32)) {
			results = append(results, fmt.Sprintf("%v", int32(field.(float32))))
			fields[i] = int32(field.(float32))
		} else {
			results = append(results, fmt.Sprintf("%v", field))
		}
	}
	return strings.Join(results, ":")
}
