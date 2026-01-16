package exercises

import "time"

// Exercise represents a named movement
type Exercise struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}
