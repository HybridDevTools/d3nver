package providers

// TypeTesting used to discriminate between other implementations
var TypeTesting = "testing"

// Testing implementation
type Testing struct {
}

// Init VM
func (t *Testing) Init() (err error) { return }

// Start VM
func (t *Testing) Start() (err error) { return }

//Stop VM
func (t *Testing) Stop() (err error) { return }

//Unregister VM
func (t *Testing) Unregister() (err error) { return }

//Update VM
func (t *Testing) Update() (updated bool, err error) { return }

// GetState VM
func (t *Testing) GetState() (state *State) { return }

// CheckIsUpdated VM
func (t *Testing) CheckIsUpdated() (updated bool, err error) { return }

//AddPostStartAction VM
func (t *Testing) AddPostStartAction(func() error) { return }

//AddPreStopAction VM
func (t *Testing) AddPreStopAction(func() error)             { return }
func (t *Testing) checkIfRunning() (running bool, err error) { return }
func (t *Testing) setState(state *State) (err error)         { return }
