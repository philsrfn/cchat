package db

import (
	"time"
)

// CurrentTime returns the current time in UTC
func CurrentTime() time.Time {
	return time.Now().UTC()
} 