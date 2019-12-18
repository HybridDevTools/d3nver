package providers

import (
	"context"
	"denver/pkg/providers/virtualbox"
	"denver/pkg/storage/http"
	"denver/pkg/util/compressor"
	"denver/pkg/util/executor"
	"denver/structs"
	"fmt"
	"path/filepath"
)

// VMProvider interface represents a VM provider
type VMProvider interface {
	Init() error
	Start() error
	Stop() error
	Unregister() error
	Update() (bool, error)
	CheckIsUpdated() (bool, error)
	GetState() *State
	AddPostStartAction(func() error)
	AddPreStopAction(func() error)
	checkIfRunning() (bool, error)
	setState(state *State) error
}

// State : Health of the virtual machine
type State struct {
	Live            bool
	OsReady         bool
	AllSystemsReady bool
}

// VMUpdater interface complements VMProvider and allow update VM components
type VMUpdater interface {
	CheckIsUpdated() (bool, error)
	Update() error
}

// NewState returns a pointer to state
func NewState() *State {
	return &State{
		Live:            false,
		OsReady:         false,
		AllSystemsReady: false,
	}
}

// GetVMProvider returns an implementation of VMProvider
func GetVMProvider(
	ctx context.Context,
	provider structs.Provider,
	instance *structs.InstanceConf,
	userDataSize int,
	channel string,
	rbiurl string,
	workingDirectory string,
) (VMProvider, error) {
	switch provider.Hypervisor {
	case TypeVirtualbox:
		updater, boxPath, err := getVMUpdater(ctx, workingDirectory, channel, rbiurl, provider.Hypervisor)
		if err != nil {
			return nil, err
		}
		return newVirtualBox(
			provider,
			instance,
			boxPath,
			executor.NewExecutor(),
			userDataSize,
			updater,
			workingDirectory,
		), nil
	}

	return nil, fmt.Errorf("invalid provider %s", provider.Hypervisor)
}

func getVMUpdater(ctx context.Context, workingDirectory, channel, rbiurl, hypervisor string) (VMUpdater, string, error) {
	switch hypervisor {
	case TypeVirtualbox:
		relBoxPath := filepath.Join("store", channel, "box.vdi")
		absBoxPath := filepath.Join(workingDirectory, relBoxPath)
		return virtualbox.NewUpdater(
			ctx,
			workingDirectory,
			fmt.Sprintf("%s/%s/virtualbox/manifest.json", rbiurl, channel),
			filepath.Join("store", channel, "manifest.json"),
			fmt.Sprintf("%s/%s/virtualbox/box.vdi.bz2", rbiurl, channel),
			relBoxPath,
			&http.HTTP{},
			compressor.NewMultiCompressor(),
		), absBoxPath, nil
	}

	return nil, "", fmt.Errorf("invalid provider %s", hypervisor)
}
