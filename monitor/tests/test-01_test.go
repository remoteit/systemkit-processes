package main

import (
	"fmt"
	"testing"
	"time"

	logging "github.com/codemodify/systemkit-logging"
	loggingC "github.com/codemodify/systemkit-logging/contracts"
	loggingP "github.com/codemodify/systemkit-logging/persisters"

	"github.com/codemodify/systemkit-processes/contracts"
	procMon "github.com/codemodify/systemkit-processes/monitor"
)

func Test_01(t *testing.T) {
	const logId = "Test_01"

	logging.Init(logging.NewEasyLoggerForLogger(loggingP.NewFileLogger(loggingC.TypeDebug, "log1.log")))

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"message": fmt.Sprintf("%s: START", logId),
	})

	processID := "test-id"

	monitor := procMon.New()
	monitor.Spawn(processID, contracts.ProcessTemplate{
		// Executable: "htop",
		Executable: "sh",
		Args:       []string{"-c", "while :; do echo 'Hit CTRL+C'; echo aaaaaaa 1>&2; sleep 1; done"},
		OnStdOut: func(data []byte) {
			logging.Instance().LogDebugWithFields(loggingC.Fields{
				"message": fmt.Sprintf("%s: OnStdOut: %v", logId, string(data)),
			})
		},
		OnStdErr: func(data []byte) {
			logging.Instance().LogDebugWithFields(loggingC.Fields{
				"message": fmt.Sprintf("%s: OnStdErr: %v", logId, string(data)),
			})
		},
	})

	logging.Instance().LogInfoWithFields(loggingC.Fields{
		"message": fmt.Sprintf(
			"%s: IsRunning: %v, ExitCode: %v, StartedAt: %v, StoppedAt: %v",
			logId,
			monitor.GetRuningProcess(processID).IsRunning(),
			monitor.GetRuningProcess(processID).ExitCode(),
			monitor.GetRuningProcess(processID).StartedAt(),
			monitor.GetRuningProcess(processID).StoppedAt(),
		),
	})

	time.Sleep(5 * time.Second)

	// stop
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"message": fmt.Sprintf("%s: STOP", logId),
	})

	monitor.Stop(processID)

	logging.Instance().LogInfoWithFields(loggingC.Fields{
		"message": fmt.Sprintf(
			"%s: IsRunning: %v, ExitCode: %v, StartedAt: %v, StoppedAt: %v",
			logId,
			monitor.GetRuningProcess(processID).IsRunning(),
			monitor.GetRuningProcess(processID).ExitCode(),
			monitor.GetRuningProcess(processID).StartedAt(),
			monitor.GetRuningProcess(processID).StoppedAt(),
		),
	})
}
