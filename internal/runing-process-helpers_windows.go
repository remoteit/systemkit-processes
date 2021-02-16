// +build windows

package internal

import (
	"golang.org/x/sys/windows"
)

var procAttrs = &windows.SysProcAttr{}
