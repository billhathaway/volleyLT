package volley

import (
	"time"
)

type (
	Request struct {
		Concurrency       int
		MaxErrors         int //not used yet
		Url               string
		Requests          int
		DisableKeepAlives bool
	}

	SessionResponse struct {
		StartTime  time.Time
		Duration   time.Duration
		Error      error
		ErrorCount int
		Responses  []Response
	}

	Response struct {
		Url        string
		StartTime  time.Time
		Duration   time.Duration
		Bytes      int
		StatusCode int
		Error      error
	}
)
