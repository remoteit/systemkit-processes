// +build !windows

package internal

import (
	"golang.org/x/sys/unix"
)

var cmdStart = []string{}
var procAttrs = &unix.SysProcAttr{}
