// Package sensu describes the payload accepted by the Sensu results API.
package sensu

// CheckResult is the JSON body posted to the Sensu results API.
// See https://docs.sensu.io/sensu-core/latest/api/results/#results-post
type CheckResult struct {
	Source      string  `json:"source"`      // source, usually the server name ($HOSTNAME)
	Name        string  `json:"name"`        // unique check name
	Output      string  `json:"output"`      // error text, or "All OK"
	Status      int     `json:"status"`      // 0 = OK, 2 = critical
	Duration    float64 `json:"duration"`    // check duration in seconds
	Occurrences int     `json:"occurrences"` // occurrences before alerting (default 3)
}
