// Package result holds the outcome of a single HTTP check.
package result

import (
	"time"
)

// Result is the outcome of one check against one domain/protocol.
type Result struct {
	IP       string
	Name     string
	Date     time.Time
	Message  string
	Duration time.Duration
	IsError  bool
}

// New builds a Result stamped with the current time.
func New(ip, name, msg string, duration time.Duration, isError bool) Result {
	return Result{
		IP:       ip,
		Name:     name,
		Date:     time.Now(),
		Message:  msg,
		Duration: duration,
		IsError:  isError,
	}
}

// String renders the result as a single log line.
func (r Result) String() string {
	return r.Date.Format(time.RFC3339) + " - " + r.Name + " - " +
		r.Duration.Round(time.Millisecond).String() + " - " + r.Message
}
