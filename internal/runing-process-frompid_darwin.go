// +build darwin

package internal

import (
	"os"

	"github.com/remoteit/systemkit-processes/contracts"
)

func getAllRuningProcesses() ([]contracts.RuningProcess, error) {
	runtimeProcesses, err := listAllRuntimeProcesses()
	if err != nil {
		return []contracts.RuningProcess{}, err
	}

	results := []contracts.RuningProcess{}
	for _, runtimeProcess := range runtimeProcesses {
		osProcess, err := os.FindProcess(runtimeProcess.ProcessID)
		if err != nil {
			continue
		}

		runingProcess := NewRuningProcessWithOSProc(
			contracts.ProcessTemplate{
				Executable:       runtimeProcess.Executable,
				Args:             runtimeProcess.Args,
				WorkingDirectory: runtimeProcess.WorkingDirectory,
				Environment:      runtimeProcess.Environment,
			},
			osProcess,
		)

		results = append(results, runingProcess)
	}

	return results, nil
}

func getRuntimeProcessByPID(pid int) (contracts.RuntimeProcess, error) {
	rps, err := listAllRuntimeProcesses()
	if err != nil {
		return contracts.RuntimeProcess{}, err
	}

	for _, rp := range rps {
		if rp.ProcessID == pid {
			return rp, nil
		}
	}

	return contracts.RuntimeProcess{}, nil
}
