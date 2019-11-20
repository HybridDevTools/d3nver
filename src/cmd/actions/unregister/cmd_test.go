package unregister

import (
	"denver/pkg/providers"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testingVM struct {
	providers.Testing

	unregisterErr error
	state         *providers.State
}

func (t *testingVM) Unregister() (err error)            { return t.unregisterErr }
func (t *testingVM) GetState() (state *providers.State) { return t.state }

func getVMProvider(provider interface{}) providers.VMProvider {
	return provider.(providers.VMProvider)
}

func TestCommandFailsIfVMIsOn(t *testing.T) {
	assert := assert.New(t)
	vm := getVMProvider(&testingVM{
		state: &providers.State{
			Live: true,
		},
	})

	cmd := NewUnregister(&vm, (*log.Logger)(nil))

	err := cmd.GetCommand().Exec()
	assert.EqualError(err, "VM is started, stop it first")
}

func TestCommandFailsIfUnregisterFails(t *testing.T) {
	assert := assert.New(t)
	vm := getVMProvider(&testingVM{
		unregisterErr: fmt.Errorf("this is fine"),
		state: &providers.State{
			Live: false,
		},
	})
	cmd := NewUnregister(&vm, (*log.Logger)(nil))

	err := cmd.GetCommand().Exec()
	assert.EqualError(err, "this is fine")
}
