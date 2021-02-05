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
		StdoutReader: func(params interface{}, outputData []byte) {
			logging.Debugf("\n%s: OnStdOut: %v", logID, string(outputData))
		},
		StderrReader: func(params interface{}, outputData []byte) {
			logging.Debugf("\n%s: OnStdErr: %v", logID, string(outputData))
		},
	}, processTag)

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

//

func TestSpawnUnix2(t *testing.T) {
	const logID = "TestSpawnUnix"

	logging.SetLogger(logging.NewStdoutLogger())
	logging.Debugf("%s: START", logID)

	processTag := "tag"
	monitor := procMon.New()
	monitor.SpawnWithTag(contracts.ProcessTemplate{
		Executable: "/usr/bin/connectd",
		Args: []string{
			"-s", "-mfg", "33024", "-ptf", "769", "-p", "bmljb2xhZUByZW1vdGUuaXQ=", "EA5AE177DCAB4A7329A853B84DF7689D6B8602EE", "80:00:00:00:01:09:D9:F4", "T33001", "2", "127.0.0.1", "0.0.0.0", "35", "0", "0",
		},
		StdoutReader: func(params interface{}, outputData []byte) {
			go func() {
				logLine := string(outputData)
				logging.Debugf("\n%s: OnStdOut: %v", logID, logLine)
			}()
		},
		StderrReader: func(params interface{}, outputData []byte) {
			logLine := string(outputData)
			logging.Debugf("\n%s: OnStdErr: %v", logID, logLine)
		},
	}, processTag)

	time.Sleep(1 * time.Hour)
}
