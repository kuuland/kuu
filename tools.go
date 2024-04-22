package kuu

import (
	"time"
)

func StartOfDay(t ...time.Time) time.Time {
	var currentTime time.Time
	if len(t) > 0 {
		currentTime = t[0]
	} else {
		currentTime = time.Now()
	}
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	return startTime
}

func EndOfDay(t ...time.Time) time.Time {
	var currentTime time.Time
	if len(t) > 0 {
		currentTime = t[0]
	} else {
		currentTime = time.Now()
	}
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location())
	return endTime
}

func hideMobile(mobile string) string {
	if mobile == "" || len(mobile) < 4 {
		return mobile
	}
	return mobile[0:3] + "****" + mobile[len(mobile)-4:]
}
