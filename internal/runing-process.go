package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	logging "github.com/remoteit/systemkit-logging"
	"github.com/remoteit/systemkit-processes/contracts"
	"github.com/remoteit/systemkit-processes/helpers"
)

const logID = "PROCESS"

// processDoesNotExist -
const processDoesNotExist = -1

type runingProcess struct {
	processTemplate    contracts.ProcessTemplate
	osCmd              *exec.Cmd
	startedAt          time.Time
	stoppedAt          time.Time
	houseKeepingMutex  *sync.Mutex
	onStoppedWasCalled bool
	isEmptyProcess     bool
}

func newRuningProcess(processTemplate contracts.ProcessTemplate, isEmptyProcess bool) *runingProcess {
	return &runingProcess{
		processTemplate:    processTemplate,
		osCmd:              nil,
		startedAt:          time.Unix(0, 0),
		stoppedAt:          time.Unix(0, 0),
		houseKeepingMutex:  &sync.Mutex{},
		onStoppedWasCalled: false,
		isEmptyProcess:     isEmptyProcess,
	}
}

// NewEmptyRuningProcess -
func NewEmptyRuningProcess() contracts.RuningProcess {
	return newRuningProcess(contracts.ProcessTemplate{}, true)
}

// NewRuningProcess -
func NewRuningProcess(processTemplate contracts.ProcessTemplate) contracts.RuningProcess {
	return newRuningProcess(processTemplate, false)
}

// NewRuningProcessWithOSProc -
func NewRuningProcessWithOSProc(processTemplate contracts.ProcessTemplate, osProc *os.Process) contracts.RuningProcess {
	rp := newRuningProcess(processTemplate, false)
	rp.osCmd = exec.Command(processTemplate.Executable, processTemplate.Args...)
	rp.osCmd.Process = osProc

	return rp
}

// Start -
func (thisRef *runingProcess) Start() error {
	if thisRef.isEmptyProcess {
		return nil
	}

	thisRef.osCmd = exec.Command(thisRef.processTemplate.Executable, thisRef.processTemplate.Args...)

	// set working folder
	if !helpers.IsNullOrEmpty(thisRef.processTemplate.WorkingDirectory) {
		thisRef.osCmd.Dir = thisRef.processTemplate.WorkingDirectory
	}

	// set env
	if thisRef.processTemplate.Environment != nil {
		thisRef.osCmd.Env = thisRef.processTemplate.Environment
	}

	// capture STDOUT
	stdOutPipe, err := thisRef.osCmd.StdoutPipe()
	if err != nil {
		logging.Errorf("%s: get-StdOut-FAIL for [%s], [%s]", logID, thisRef.processTemplate.Executable, err.Error())
		return err
	}

	go func() {
		thisRef.osCmd.Process.Wait() // read the exit code on quit

		logging.Debugf("%s: read-STDOUT for [%s]", logID, thisRef.processTemplate.Executable)
		err := readOutput(stdOutPipe, thisRef.processTemplate.StdoutReader, thisRef.processTemplate.StdoutReaderParams)
		if err != nil {
			logging.Warningf("%s: read-STDOUT-FAIL for [%s], [%s]", logID, thisRef.processTemplate.Executable, err.Error())
		}
		logging.Debugf("%s: read-STDOUT-SUCESS for [%s]", logID, thisRef.processTemplate.Executable)

		if thisRef.processTemplate.StdoutReader != nil {
			thisRef.houseKeepingMutex.Lock()
			defer thisRef.houseKeepingMutex.Unlock()

			if !thisRef.onStoppedWasCalled {
				if thisRef.processTemplate.OnStopped != nil {
					thisRef.processTemplate.OnStopped(thisRef.processTemplate.OnStoppedParams)
				}

				thisRef.onStoppedWasCalled = true
			}
		}
	}()

	// capture STDERR
	if thisRef.processTemplate.StderrReader != nil {
		stdErrPipe, err := thisRef.osCmd.StderrPipe()
		if err != nil {
			logging.Errorf("%s: get-StdErr-FAIL for [%s], [%s]", logID, thisRef.processTemplate.Executable, err.Error())
			return err
		}

		go func() {
			logging.Debugf("%s: read-STDERR for [%s]", logID, thisRef.processTemplate.Executable)
			err := readOutput(stdErrPipe, thisRef.processTemplate.StderrReader, thisRef.processTemplate.StderrReaderParams)
			if err != nil {
				logging.Warningf("%s: read-STDERR-FAIL for [%s], [%s]", logID, thisRef.processTemplate.Executable, err.Error())
			}
			logging.Debugf("%s: read-STDERR-SUCESS for [%s]", logID, thisRef.processTemplate.Executable)

			//
			thisRef.houseKeepingMutex.Lock()
			defer thisRef.houseKeepingMutex.Unlock()

			if !thisRef.onStoppedWasCalled {
				if thisRef.processTemplate.OnStopped != nil {
					thisRef.processTemplate.OnStopped(thisRef.processTemplate.OnStoppedParams)
				}

				thisRef.onStoppedWasCalled = true
			}
		}()
	}

	thisRef.osCmd.SysProcAttr = procAttrs

	// start
	logging.Debugf("%s: start %s", logID, helpers.AsJSONString(thisRef.processTemplate))

	err := thisRef.osCmd.Start()
	if err != nil {
		thisRef.stoppedAt = time.Now()

		detailedErr := fmt.Errorf("%s: start-FAILED %s, %s", logID, helpers.AsJSONString(thisRef.processTemplate), err.Error())
		logging.Error(detailedErr.Error())

		return detailedErr
	}

	thisRef.startedAt = time.Now()

	return nil
}

// Stop - stops the process
func (thisRef *runingProcess) Stop(tag string, attempts int, waitTimeout time.Duration) error {
	if thisRef.osCmd == nil || thisRef.osCmd.Process == nil {
		return nil
	}

	if !thisRef.IsRunning() {
		thisRef.osCmd.Process.Wait()
		return nil
	}

	defer func() {
		logging.Debugf("%s: STOP-END %s", logID, tag)
	}()

	logging.Debugf("%s: STOP-START %s", logID, tag)

	var err error
	count := 0
	maxStopAttempts := 20
	for {
		// try #
		count++
		if count > maxStopAttempts {
			logging.Errorf("%s: stop-FAIL [%s] with PID [%d]", logID, thisRef.processTemplate.Executable, thisRef.processID())
			break
		}

		// log the STOP attempt #

		for i := 0; i < attempts; i++ {
			logging.Debugf("%s: stop-ATTEMPT-SIGINT #%d to stop [%s]", logID, i, thisRef.processTemplate.Executable)
			thisRef.osCmd.Process.Signal(syscall.SIGINT) // this works on all except on Windows
			sendCtrlC(thisRef.osCmd.Process.Pid)         // this works on Windows
			time.Sleep(waitTimeout)

			thisRef.osCmd.Process.Wait()
			if !thisRef.IsRunning() {
				thisRef.stoppedAt = time.Now()
				logging.Debugf("%s: stop-SUCCESS [%s]", logID, thisRef.processTemplate.Executable)
				return nil
			}
		}

		for i := 0; i < attempts; i++ {
			logging.Debugf("%s: stop-ATTEMPT-SIGTERM #%d to stop [%s]", logID, i, thisRef.processTemplate.Executable)
			thisRef.osCmd.Process.Signal(syscall.SIGTERM)
			time.Sleep(waitTimeout)

			thisRef.osCmd.Process.Wait()
			if !thisRef.IsRunning() {
				thisRef.stoppedAt = time.Now()
				logging.Debugf("%s: stop-SUCCESS [%s]", logID, thisRef.processTemplate.Executable)
				return nil
			}
		}

		for i := 0; i < attempts; i++ {
			logging.Debugf("%s: stop-ATTEMPT-SIGKILL #%d to stop [%s]", logID, i, thisRef.processTemplate.Executable)
			thisRef.osCmd.Process.Signal(syscall.SIGKILL)
			time.Sleep(waitTimeout)

			thisRef.osCmd.Process.Wait()
			if !thisRef.IsRunning() {
				thisRef.stoppedAt = time.Now()
				logging.Debugf("%s: stop-SUCCESS [%s]", logID, thisRef.processTemplate.Executable)
				return nil
			}
		}

		for i := 0; i < attempts; i++ {
			logging.Debugf("%s: stop-ATTEMPT-aggressive-kill-1 #%d to stop [%s]", logID, i, thisRef.processTemplate.Executable)
			processKillHelper(thisRef.osCmd.Process.Pid)
			time.Sleep(waitTimeout)

			thisRef.osCmd.Process.Wait()
			if !thisRef.IsRunning() {
				thisRef.stoppedAt = time.Now()
				logging.Debugf("%s: stop-SUCCESS [%s]", logID, thisRef.processTemplate.Executable)
				return nil
			}
		}

		for i := 0; i < attempts; i++ {
			logging.Debugf("%s: stop-ATTEMPT-aggressive-kill-2 #%d to stop [%s]", logID, i, thisRef.processTemplate.Executable)
			err = thisRef.osCmd.Process.Kill()
			time.Sleep(waitTimeout)

			thisRef.osCmd.Process.Wait()
			if !thisRef.IsRunning() {
				thisRef.stoppedAt = time.Now()
				logging.Debugf("%s: stop-SUCCESS [%s]", logID, thisRef.processTemplate.Executable)
				return nil
			}
		}
	}

	return err
}

// IsRunning - tells if the process is running
func (thisRef *runingProcess) IsRunning() bool {
	if thisRef.osCmd == nil || thisRef.osCmd.Process == nil {
		return false
	}

	pid := thisRef.processID()
	if pid == processDoesNotExist {
		return false
	}

	rp := thisRef.Details()

	return (rp.State != contracts.ProcessStateNonExistent &&
		rp.State != contracts.ProcessStateObsolete &&
		rp.State != contracts.ProcessStateDead &&
		rp.State != contracts.ProcessStateUnknown)
}

// Details - return processTemplate about the process
func (thisRef *runingProcess) Details() contracts.RuntimeProcess {
	if thisRef.osCmd == nil || thisRef.osCmd.Process == nil {
		return contracts.RuntimeProcess{
			State: contracts.ProcessStateNonExistent,
		}
	}

	rpByPID, err := getRuntimeProcessByPID(thisRef.processID())
	if err != nil {
		return contracts.RuntimeProcess{
			State: contracts.ProcessStateNonExistent,
		}
	}

	return rpByPID
}

// ExitCode -
func (thisRef *runingProcess) ExitCode() int {
	if thisRef.osCmd == nil || thisRef.osCmd.Process == nil || thisRef.osCmd.ProcessState == nil {
		return 0
	}

	return thisRef.osCmd.ProcessState.ExitCode()
}

// StartedAt - returns the time when the process was started
func (thisRef *runingProcess) StartedAt() time.Time {
	if thisRef.osCmd == nil || thisRef.osCmd.Process == nil {
		return time.Unix(0, 0)
	}

	return thisRef.startedAt
}

// StoppedAt - returns the time when the process was stopped
func (thisRef *runingProcess) StoppedAt() time.Time {
	if thisRef.osCmd == nil || thisRef.osCmd.Process == nil {
		return time.Unix(0, 0)
	}

	return thisRef.stoppedAt
}

func (thisRef *runingProcess) processID() int {
	if thisRef.osCmd == nil || thisRef.osCmd.Process == nil {
		return processDoesNotExist
	}

	return thisRef.osCmd.Process.Pid
}

func readOutput(readerCloser io.ReadCloser, outputReader contracts.ProcessOutputReader, params interface{}) error {
	reader := bufio.NewReader(readerCloser)
	line, _, err := reader.ReadLine()
	for {
		if err != nil {
			break
		}

		if outputReader != nil {
			outputReader(params, line)
		}

		line, _, err = reader.ReadLine()
	}

	readerCloser.Close()

	if err == io.EOF {
		return nil
	}

	return err
}
