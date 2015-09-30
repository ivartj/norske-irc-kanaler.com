package main

import (
	"time"
	"fmt"
	"math"
)

func timeAgo(t time.Time) string {
	dur := time.Since(t)
	days := int(math.Floor(dur.Hours() / 24))
	hours := int(math.Floor(dur.Hours()))
	minutes := int(math.Floor(dur.Minutes()))

	switch {
	case days == 1:
		return "1 dag siden"
	case days > 1:
		return fmt.Sprintf("%d dager siden", days)
	case hours == 1:
		return "1 time siden"
	case hours > 1:
		return fmt.Sprintf("%d timer siden", hours)
	case minutes == 1:
		return "1 minutt siden"
	case minutes > 1:
		return fmt.Sprintf("%d minutter siden", minutes)
	default:
		return "innen 1 minutt siden"
	}
}

