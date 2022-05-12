package util

import (
	"strconv"
	"time"
)

func HoursMinutes(dt time.Time) int {
	minutes := strconv.Itoa(dt.Minute())
	if dt.Minute() == 0 {
		minutes = "00"
	} else if dt.Minute() < 10 {
		minutes = "0" + minutes
	}
	hour := strconv.Itoa(dt.Hour())
	if dt.Hour() == 0 {
		hour = ""
	}
	hourMinute, _ := strconv.Atoi(hour + minutes)
	return hourMinute
}
