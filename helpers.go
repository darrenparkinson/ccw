package ccw

import (
	"errors"
	"math"
	"regexp"
	"strconv"
)

func isoDurationToMonthsFloat(isoDuration string) (float64, error) {
	re := regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:.\d+)?)S)?$`)
	matches := re.FindStringSubmatch(isoDuration)
	if matches == nil {
		return 0, errors.New("input string is of incorrect format")
	}

	months := 0.0

	// we should only get months and days, so we'll ignore hours/minutes/seconds

	// months
	if matches[2] != "" {
		f, err := strconv.ParseFloat(matches[2], 32)
		if err != nil {
			return 0, err
		}
		months += (f)
	}

	// days
	if matches[3] != "" {
		f, err := strconv.ParseFloat(matches[3], 32)
		if err != nil {
			return 0, err
		}
		months += (f / 30)
	}

	return math.Round(months*100) / 100, nil
}

func isoDurationToDaysFloat(isoDuration string) (float64, error) {
	re := regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:.\d+)?)S)?$`)
	matches := re.FindStringSubmatch(isoDuration)
	if matches == nil {
		return 0, errors.New("input string is of incorrect format")
	}

	days := 0.0

	// Cisco appears to only put a value in days

	// days
	if matches[3] != "" {
		f, err := strconv.ParseFloat(matches[3], 32)
		if err != nil {
			return 0, err
		}
		days += (f)
	}

	return math.Round(days*100) / 100, nil
}
