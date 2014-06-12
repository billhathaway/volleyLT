// statsReporter.go
package main

import (
	"bytes"
	"fmt"
	"github.com/billhathaway/volleyLT/common"
)

type (
	statReporter struct{}
)

func (sr *statReporter) report(responses []*volley.SessionResponse) string {
	statusCodes := make(map[int]int)
	var totalResponseTimeMs int64
	var totalBytes int
	var maxBytes int
	var maxResponseTimeMs int64
	var responseTimeMs int64
	var responseCount int64
	for _, sr := range responses {
		for _, response := range sr.Responses {
			responseCount++
			statusCodes[response.StatusCode]++
			responseTimeMs = response.Duration.Nanoseconds() / 1000000
			totalResponseTimeMs += responseTimeMs
			if responseTimeMs > maxResponseTimeMs {
				maxResponseTimeMs = responseTimeMs
			}
			totalBytes += response.Bytes
			if response.Bytes > maxBytes {
				maxBytes = response.Bytes
			}

		}
	}
	buf := bytes.Buffer{}
	buf.WriteString("statusCodes: ")
	for code, count := range statusCodes {
		buf.WriteString(fmt.Sprintf("%d=%d ", code, count))
	}
	buf.WriteString(fmt.Sprintf("\navgResponseTimeMs=%d maxResponseTimeMs=%d\n", totalResponseTimeMs/responseCount, maxResponseTimeMs))
	return buf.String()
}
