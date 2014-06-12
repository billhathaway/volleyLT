// reporter
package main

import (
	"github.com/billhathaway/volleyLT/common"
)

type (
	reporter interface {
		report(*volley.SessionResponse) string
	}
)
