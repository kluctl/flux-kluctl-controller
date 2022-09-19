package controllers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var weekdays = []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}

var timeSpecPattern = regexp.MustCompile(`^([a-zA-Z]{3})-([a-zA-Z]{3}) (\d\d):(\d\d)-(\d\d):(\d\d) (?P<tz>[a-zA-Z/_]+)$`)
var iso8601TimeSpecPattern = `(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2})`
var absoluteTimeSpecPattern = regexp.MustCompile(fmt.Sprintf("^%s-%s$", iso8601TimeSpecPattern, iso8601TimeSpecPattern))

func MatchesTimeSpec(t time.Time, spec string) (bool, error) {
	if strings.ToLower(spec) == "always" {
		return true, nil
	} else if strings.ToLower(spec) == "never" {
		return false, nil
	}

	for _, spec_ := range strings.Split(spec, ",") {
		spec_ = strings.TrimSpace(spec_)
		recurringMatch := timeSpecPattern.FindStringSubmatch(spec_)
		if recurringMatch != nil {
			hit, err := matchesRecurringTimeSpec(t, recurringMatch)
			if err != nil {
				return false, err
			}
			if hit {
				return true, nil
			}
		}

		absolueMatch := absoluteTimeSpecPattern.FindStringSubmatch(spec_)
		if recurringMatch != nil {
			hit, err := matchedAbsoluteTimeSpec(t, absolueMatch)
			if err != nil {
				return false, err
			}
			if hit {
				return true, nil
			}
		}
		if recurringMatch == nil && absolueMatch == nil {
			return false, fmt.Errorf(`time spec value "%s" does not match format ("Mon-Fri 06:30-20:30 Europe/Berlin" or "2019-01-01T00:00:00+00:00-2019-01-02T12:34:56+00:00")`, spec_)
		}
	}
	return false, nil
}

func weekdayIndex(s string) int {
	s = strings.ToLower(s)
	for i, x := range weekdays {
		if x == s {
			return i
		}
	}
	return -1
}

func matchesRecurringTimeSpec(t time.Time, match []string) (bool, error) {
	calcMinutes := func(i1, i2 int) (int, error) {
		h, err := strconv.ParseInt(match[i1], 10, 32)
		if err != nil {
			return 0, err
		}
		m, err := strconv.ParseInt(match[i2], 10, 32)
		if err != nil {
			return 0, err
		}
		return int(h*60 + m), nil
	}

	localTime := t
	tz := match[timeSpecPattern.SubexpIndex("tz")]
	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return false, err
		}
		localTime = t.In(loc)
	}

	localWeekday := int(localTime.Weekday())
	if localWeekday == 0 {
		localWeekday = 6
	} else {
		localWeekday--
	}

	dayFrom := weekdayIndex(match[1])
	dayTo := weekdayIndex(match[2])
	if dayFrom == -1 {
		return false, fmt.Errorf("invalid day (%s) specified", match[1])
	}
	if dayTo == -1 {
		return false, fmt.Errorf("invalid day (%s) specified", match[2])
	}

	dayMatches := false
	if dayFrom > dayTo {
		// wrap around, e.g. Sun-Fri (makes sense for countries with work week starting on Sunday)
		dayMatches = localWeekday >= dayFrom || localWeekday <= dayTo
	} else {
		// e.g. Mon-Fri
		dayMatches = dayFrom <= localWeekday && localWeekday <= dayTo
	}
	localTimeMinutes := localTime.Hour()*60 + localTime.Minute()
	minuteFrom, err := calcMinutes(3, 4)
	if err != nil {
		return false, err
	}
	minuteTo, err := calcMinutes(5, 6)
	if err != nil {
		return false, err
	}
	timeMatches := localTimeMinutes >= minuteFrom && localTimeMinutes <= minuteTo
	return dayMatches && timeMatches, nil
}

func matchedAbsoluteTimeSpec(t time.Time, match []string) (bool, error) {
	timeFrom, err := time.Parse(time.RFC3339, match[1])
	if err != nil {
		return false, err
	}
	timeTo, err := time.Parse(time.RFC3339, match[2])
	if err != nil {
		return false, err
	}
	tu := t.UnixMilli()
	tfu := timeFrom.UnixMilli()
	ttu := timeTo.UnixMilli()
	return tfu <= tu && tu <= ttu, nil
}
