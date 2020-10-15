package find

import (
	"github.com/remoteit/systemkit-processes/contracts"
	"github.com/remoteit/systemkit-processes/internal"
)

// ProcessByPID - finds process by PID
func ProcessByPID(pid int) (contracts.RuningProcess, error) {
	return internal.GetRuningProcessByPID(pid)
}

// AllProcesses - returns all processes
func AllProcesses() ([]contracts.RuningProcess, error) {
	return internal.GetAllRuningProcesses()
}
