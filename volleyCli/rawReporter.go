// reporter
package main

import (
	"bytes"
	"fmt"
	"github.com/billhathaway/volleyLT/common"
)

type (
	rawReporter struct{}
)

func (r *rawReporter) report(vr *volley.SessionResponse) string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("ErrorCount: %d\n", vr.ErrorCount))
	for _, response := range vr.Responses {
		if response.Error == nil {
			buf.WriteString(fmt.Sprintf("code=%d startTime=%d ms=%d len=%d\n", response.StatusCode, response.StartTime.UnixNano(), response.Duration.Nanoseconds()/1000000, response.Bytes))
		} else {
			buf.WriteString(fmt.Sprintf("code=%d startTime=%d ms=%d len=%d err=%q\n", response.StatusCode, response.StartTime.UnixNano(), response.Duration.Nanoseconds()/1000000, response.Bytes, response.Error.Error()))
		}

	}
	return buf.String()
}
