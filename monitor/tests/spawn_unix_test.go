// +build !windows

package tests

import (
	"testing"
	"time"

	logging "github.com/remoteit/systemkit-logging"

	"github.com/remoteit/systemkit-processes/contracts"
	procMon "github.com/remoteit/systemkit-processes/monitor"
)

func TestSpawnUnix(t *testing.T) {
	const logID = "TestSpawnUnix"

	logging.SetLogger(logging.NewStdoutLogger())
	logging.Debugf("%s: START", logID)

	processTag := "aaaaaaaaa"
	monitor := procMon.New()
	monitor.SpawnWithTag(contracts.ProcessTemplate{
		Executable: "sh",
		Args:       []string{"-c", "while :; do echo 'Hit CTRL+C'; echo aaaaaaa 1>&2; sleep 1; done"},
	}, processTag)
	monitor.GetProcess(processTag).OnStdOut(func(params interface{}, outputData []byte) {
		logging.Debugf("\n%s: OnStdOut: %v", logID, string(outputData))
	}, nil)
	monitor.GetProcess(processTag).OnStdErr(func(params interface{}, outputData []byte) {
		logging.Debugf("\n%s: OnStdErr: %v", logID, string(outputData))
	}, nil)

	logging.Infof(
		"%s: IsRunning: %v, ExitCode: %v, StartedAt: %v, StoppedAt: %v",
		logID,
		monitor.GetProcess(processTag).IsRunning(),
		monitor.GetProcess(processTag).ExitCode(),
		monitor.GetProcess(processTag).StartedAt(),
		monitor.GetProcess(processTag).StoppedAt(),
	)

	// WAIT 5 seconds
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				logging.Debugf("\n%s: Tick at, %v", logID, t)
			}
		}
	}()
	time.Sleep(5 * time.Second)
	ticker.Stop()
	done <- true

	// STOP
	logging.Debugf("%s: STOP", logID)

	monitor.Stop(processTag)

	logging.Infof(
		"%s: IsRunning: %v, ExitCode: %v, StartedAt: %v, StoppedAt: %v",
		logID,
		monitor.GetProcess(processTag).IsRunning(),
		monitor.GetProcess(processTag).ExitCode(),
		monitor.GetProcess(processTag).StartedAt(),
		monitor.GetProcess(processTag).StoppedAt(),
	)
}
