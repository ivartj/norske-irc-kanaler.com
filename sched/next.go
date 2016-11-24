package sched

import (
	"time"
	"fmt"
	"regexp"
	"strconv"
)

// Takes a time string in one of the following forms:
//
//  * 15:04
//
//  * Monday 15:04
//
// Returns the next time that matches the given specification.
func Next(str string) (time.Time, error) {
	t, err := NextFollowing(str, time.Now())
	return t, err
}

var regexWeekdayClock = regexp.MustCompile(`^(.+?) ([0-9]{2}):([0-9]{2})`)
var regexClock = regexp.MustCompile(`([0-9]{2}):([0-9]{2})`)
var weekdays = map[string]time.Weekday{
	"Sunday": time.Sunday,
	"Monday": time.Monday,
	"Tuesday": time.Tuesday,
	"Wednesday": time.Wednesday,
	"Thursday": time.Thursday,
	"Friday": time.Friday,
	"Saturday": time.Saturday,
}

func NextFollowing(str string, start time.Time) (time.Time, error) {

	t := start

	switch {

	case regexWeekdayClock.MatchString(str):

		submatches := regexWeekdayClock.FindStringSubmatch(str)
		if len(submatches) != 4 {
			return time.Time{}, fmt.Errorf("Unexpected number of submatches in '%s'", str)
		}
		weekday, ok := weekdays[submatches[1]]
		if !ok {
			return time.Time{}, fmt.Errorf("Unrecognized weekday, '%s'", submatches[1])
		}
		hour, err := strconv.Atoi(submatches[2])
		if err != nil || hour >= 24 {
			return time.Time{}, fmt.Errorf("Invalid hour, '%s'", submatches[2])
		}
		min, err := strconv.Atoi(submatches[3])
		if err != nil || min >= 60 {
			return time.Time{}, fmt.Errorf("Invalid minute, '%s'", submatches[3])
		}
		
		dayOffset := (time.Weekday(7) + weekday - t.Weekday()) % time.Weekday(7)
		t = t.Add(time.Duration(dayOffset) * time.Hour * time.Duration(24))
		t = t.Truncate(time.Duration(24) * time.Hour)
		t = t.Add(time.Duration(hour) * time.Hour + time.Duration(min) * time.Minute)

		if t.Before(start) {
			t = t.Add(time.Duration(24 * 7) * time.Hour)
		}

		return t, nil

	case regexClock.MatchString(str):

		submatches := regexClock.FindStringSubmatch(str)
		if len(submatches) != 3 {
			return time.Time{}, fmt.Errorf("Unexpected number of submatches in '%s'", str)
		}

		hour, err := strconv.Atoi(submatches[1])
		if err != nil || hour >= 24 {
			return time.Time{}, fmt.Errorf("Invalid hour, '%s'", submatches[1])
		}
		min, err := strconv.Atoi(submatches[2])
		if err != nil || min >= 60 {
			return time.Time{}, fmt.Errorf("Invalid minute, '%s'", submatches[2])
		}

		t = t.Truncate(time.Duration(24) * time.Hour)
		t = t.Add(time.Duration(hour) * time.Hour + time.Duration(min) * time.Minute)

		if t.Before(start) {
			t = t.Add(time.Duration(24) * time.Hour)
		}

		return t, nil
	}

	return time.Time{}, fmt.Errorf("Unrecognized time format, '%s'", str)
}

