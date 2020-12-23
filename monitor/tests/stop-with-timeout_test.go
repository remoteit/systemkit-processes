package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/remoteit/systemkit-processes/contracts"
	procMon "github.com/remoteit/systemkit-processes/monitor"
)

func TestAbs(t *testing.T) {
	processTag := "aaaa"

	monitor := procMon.New()
	monitor.SpawnWithTag(contracts.ProcessTemplate{
		Executable: "C:\\Program Files\\remoteit-bin\\connectd.exe",
		Args:       []string{"-s", "-e", "bWF4X2RlcHRoIDM1CmFwcGxpY2F0aW9uX3R5cGUgMzUKcHJveHlfZGVzdF9pcCAxMjcuMC4wLjEKbWFudWZhY3R1cmVfaWQgMzMwMjQKcGxhdGZvcm1fdmVyc2lvbiA1CnByb3h5X2xvY2FsX3BvcnQgNjU1MzUKcHJveHlfZGVzdF9wb3J0IDY1NTM1CmFwcGxpY2F0aW9uX3R5cGVfb3ZlcmxvYWQgNDAKVUlEIDgwOjAwOjAwOjAwOjAxOjA5OjVFOkY1CnNlY3JldCA3NEY5MDhDRjlENEMxREYzNEJFRjZFRjU5RjM0ODYwMkIzNTQwODRGCiMK"},
	}, processTag)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		time.Sleep(30 * time.Second)
		wg.Done()
	}()

	wg.Wait()
	monitor.StopWithTimeout(processTag, 10, 500*time.Millisecond)
}
