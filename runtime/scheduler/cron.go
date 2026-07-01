package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type cronExpr struct {
	second, minute []int
	hour           []int
	day, month, wk []int
}

func parseCron(expr string) (*cronExpr, error) {
	fields := strings.Fields(expr)
	if len(fields) != 6 {
		return nil, fmt.Errorf("cron expression must have 6 fields (second minute hour day month weekday), got %d", len(fields))
	}

	c := &cronExpr{}
	var err error

	c.second, err = parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("second: %w", err)
	}
	c.minute, err = parseField(fields[1], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("minute: %w", err)
	}
	c.hour, err = parseField(fields[2], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("hour: %w", err)
	}
	c.day, err = parseField(fields[3], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("day: %w", err)
	}
	c.month, err = parseField(fields[4], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("month: %w", err)
	}
	c.wk, err = parseField(fields[5], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("weekday: %w", err)
	}

	return c, nil
}

func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		result := make([]int, 0, max-min+1)
		for i := min; i <= max; i++ {
			result = append(result, i)
		}
		return result, nil
	}

	var result []int
	seen := make(map[int]bool)

	parts := strings.Split(field, ",")
	for _, part := range parts {
		if strings.Contains(part, "/") {
			stepParts := strings.Split(part, "/")
			if len(stepParts) != 2 {
				return nil, fmt.Errorf("invalid step expression: %s", part)
			}

			step, err := strconv.Atoi(stepParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid step value: %s", stepParts[1])
			}

			rangeMin, rangeMax := min, max
			rangePart := stepParts[0]
			if rangePart != "*" && strings.Contains(rangePart, "-") {
				rp := strings.Split(rangePart, "-")
				rangeMin, err = strconv.Atoi(rp[0])
				if err != nil {
					return nil, fmt.Errorf("invalid range min: %s", rp[0])
				}
				rangeMax, err = strconv.Atoi(rp[1])
				if err != nil {
					return nil, fmt.Errorf("invalid range max: %s", rp[1])
				}
			} else if rangePart != "*" {
				rangeMin, err = strconv.Atoi(rangePart)
				if err != nil {
					return nil, fmt.Errorf("invalid value: %s", rangePart)
				}
				rangeMax = rangeMin
			}

			for i := rangeMin; i <= rangeMax; i += step {
				if i >= min && i <= max && !seen[i] {
					seen[i] = true
					result = append(result, i)
				}
			}
		} else if strings.Contains(part, "-") {
			rp := strings.Split(part, "-")
			rMin, err := strconv.Atoi(rp[0])
			if err != nil {
				return nil, fmt.Errorf("invalid range min: %s", rp[0])
			}
			rMax, err := strconv.Atoi(rp[1])
			if err != nil {
				return nil, fmt.Errorf("invalid range max: %s", rp[1])
			}
			for i := rMin; i <= rMax; i++ {
				if i >= min && i <= max && !seen[i] {
					seen[i] = true
					result = append(result, i)
				}
			}
		} else {
			val, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid value: %s", part)
			}
			if val >= min && val <= max && !seen[val] {
				seen[val] = true
				result = append(result, val)
			}
		}
	}

	if len(result) == 0 {
		for i := min; i <= max; i++ {
			result = append(result, i)
		}
	}

	return result, nil
}

func (c *cronExpr) next(from time.Time) time.Time {
	t := from.Truncate(time.Second).Add(time.Second)

	match := func(val int, list []int) bool {
		for _, v := range list {
			if v == val {
				return true
			}
		}
		return false
	}

	for i := 0; i < 525600; i++ {
		if match(int(t.Month()), c.month) &&
			match(t.Day(), c.day) &&
			match(int(t.Weekday()), c.wk) &&
			match(t.Hour(), c.hour) &&
			match(t.Minute(), c.minute) &&
			match(t.Second(), c.second) {
			return t
		}
		t = t.Add(time.Second)
	}

	return time.Time{}
}
