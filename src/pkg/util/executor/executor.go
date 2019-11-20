package executor

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"
)

// Executor with mutex
type Executor struct {
	mutex sync.Mutex
}

// NewExecutor returns a pointer to Executor
func NewExecutor() *Executor {
	return &Executor{
		mutex: sync.Mutex{},
	}
}

// Execute executes a command with arguments
func (e *Executor) Execute(args []string) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	cmd := exec.Command(args[0], args[1:]...)

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(fmt.Sprintf("%s\n%s\n%s", err.Error(), stdErr.String(), stdOut.String()))
	}

	return stdOut.String(), nil
}
