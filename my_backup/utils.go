package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func marshalTime(t time.Time) string {
	hms := strings.Split(t.Format(time.TimeOnly), ":")
	return fmt.Sprintf("/%v_%v-%v-%v", t.Format(time.DateOnly), hms[0], hms[1], hms[2])
}

func unmarshalTime(tstr string) (time.Time, error) {
	var year, month, day, hour, minute, second int
	_, err := fmt.Sscanf(tstr, "%d-%d-%d_%d-%d-%d", &year, &month, &day, &hour, &minute, &second)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, &time.Location{}), nil
}
