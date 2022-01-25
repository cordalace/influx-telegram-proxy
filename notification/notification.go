package notification

import "time"

type Notification struct {
	CheckID   string
	CheckName string
	Level     string
	Message   string
	Time      time.Time
	Type      string
}
