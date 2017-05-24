package result

import (
	"time"
)

func New(ip, name, msg string, is_error bool) Result {
	result := Result{
		Ip: ip,
		Name:    name,
		Date:    time.Now(),
		Message: msg,
		IsError: is_error,
	}
	return result
}

type Result struct {
	Ip string
	Name    string
	Date    time.Time
	Message string
	IsError bool
}

func (r Result) String() string {
	return r.Date.String() + " - " + r.Name + " - " + r.Message
}
