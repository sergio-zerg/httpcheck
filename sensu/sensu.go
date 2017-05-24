package sensu

type SensuCheckResult struct {
    Source      string  `json:"source"`  // Source (ussally server name $HOSTNAME)
    Name        string  `json:"name"`    // unique check name
    Output      string  `json:"output"`  // error text or OK
    Status      int     `json:"status"`  // 0 - if OK, else 2
    Duration    float64 `json:"duration"` // check time duration
    Occurrences int     `json:"occurrences"` // Check time before alerting. Default 3
}