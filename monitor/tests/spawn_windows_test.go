// +build windows

package tests

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	logging "github.com/remoteit/systemkit-logging"

	"github.com/remoteit/systemkit-processes/contracts"
	procMon "github.com/remoteit/systemkit-processes/monitor"
)

func TestSpawnWindows(t *testing.T) {
	const logID = "TestSpawnWindows"

	logging.Debugf("%s: START", logID)

	monitor := procMon.New()

	processTag, _ := monitor.Spawn(contracts.ProcessTemplate{
		Executable: "notepad.exe",
		Args:       []string{},
		StdoutReader: func(params interface{}, outputData []byte) {
			logging.Debugf("%s: OnStdOut: %v", logID, string(outputData))
		},
		StderrReader: func(params interface{}, outputData []byte) {
			logging.Debugf("%s: OnStdErr: %v", logID, string(outputData))
		},
	})

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
				logging.Debugf("%s: Tick at, %v", logID, t)
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

func TestSpawnWindowsWithTimeout(t *testing.T) {
	wg := sync.WaitGroup{}

	udpPortG := 0

	processTag := "aaaa"

	monitor := procMon.New()
	monitor.SpawnWithTag(contracts.ProcessTemplate{
		Executable: "C:\\connectd.exe",
		Args: []string{
			"-s", "-mfg", "33280", "-ptf", "256", "-p", "bmljb2xhZUByZW1vdGUuaXQ=", "EA5AE177DCAB4A7329A853B84DF7689D6B8602EE", "80:00:00:00:01:0A:01:BA", "T30002", "2", "1.1.1.1", "0.0.0.0", "35", "0", "0",
		},
		StdoutReader: func(params interface{}, outputData []byte) {
			rawDataLine := string(outputData)

			udpPortStr := "bound to UDP port"
			if strings.Contains(rawDataLine, udpPortStr) {

				rawDataLine = rawDataLine[strings.Index(rawDataLine, udpPortStr)+len(udpPortStr):]
				rawDataLine = strings.TrimSpace(rawDataLine)

				udpPort, err := strconv.Atoi(rawDataLine)
				if err == nil {
					udpPortG = udpPort
				}
			}
		},
		OnStopped: func(params interface{}) {
			fmt.Println("STOPPED !!!")
		},
	}, processTag)

	wg.Add(1)
	go func() {
		time.Sleep(3 * time.Second)
		wg.Done()
	}()
	wg.Wait()

	fmt.Println("SendKillPacket !!!")
	SendKillPacket(udpPortG)

	wg.Add(1)
	go func() {
		time.Sleep(3 * time.Second)
		wg.Done()
	}()
	wg.Wait()

	fmt.Println("StopWithTimeout !!!")
	monitor.StopWithTimeout(processTag, 2, 100*time.Millisecond)

	wg.Add(1)
	go func() {
		time.Sleep(3 * time.Second)
		wg.Done()
	}()
	wg.Wait()

	fmt.Println("DONE !!!")
}
