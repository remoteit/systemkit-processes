package contracts

// ProcessOutputReader -
type ProcessOutputReader func(params interface{}, outputData []byte)

// ProcessStoppedDelegate -
type ProcessStoppedDelegate func(params interface{})

// ProcessTemplate -
type ProcessTemplate struct {
	Executable         string                 `json:"executable"`
	Args               []string               `json:"args"`
	WorkingDirectory   string                 `json:"workingDirectory"`
	Environment        []string               `json:"environment"`
	StdoutReader       ProcessOutputReader    `json:"-"`
	StdoutReaderParams interface{}            `json:"-"`
	StderrReader       ProcessOutputReader    `json:"-"`
	StderrReaderParams interface{}            `json:"-"`
	OnStopped          ProcessStoppedDelegate `json:"-"`
	OnStoppedParams    interface{}            `json:"-"`
}
