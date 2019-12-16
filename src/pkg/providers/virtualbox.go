package providers

import (
	"bufio"
	"denver/pkg/util"
	"denver/pkg/util/executor"
	"denver/structs"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// TypeVirtualbox used to discriminate between other implementations
var TypeVirtualbox = "virtualbox"

// Virtualbox implementation
type Virtualbox struct {
	provider     structs.Provider
	instance     *structs.InstanceConf
	boxPath      string
	userData     string
	userDataSize int
	executor     *executor.Executor
	state        *State
	updater      VMUpdater

	postStartActions []func() error
	preStopActions   []func() error
}

func newVirtualBox(
	provider structs.Provider,
	instance *structs.InstanceConf,
	boxPath string,
	executor *executor.Executor,
	userDataSize int,
	updater VMUpdater,
	workingDirectory string,
) *Virtualbox {
	return &Virtualbox{
		provider:     provider,
		instance:     instance,
		boxPath:      boxPath,
		userData:     filepath.Join(workingDirectory, "store", "userdata.vdi"),
		userDataSize: userDataSize,
		executor:     executor,
		state:        NewState(),
		updater:      updater,
	}
}

// Init VM
func (v *Virtualbox) Init() (err error) {
	_, err = v.Update()
	if err != nil {
		return
	}

	exists, err := v.checkIfExists()
	if err != nil {
		return
	}

	if exists {
		return fmt.Errorf("%s already exists", v.instance.Name)
	}
	return v.init()
}

// Start VM
func (v *Virtualbox) Start() (err error) {
	if exists, err := v.checkIfExists(); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("%s does not exists", v.instance.Name)
	}

	if isRunning, err := v.checkIfRunning(); err != nil {
		return err
	} else if isRunning {
		return fmt.Errorf("%s is running", v.instance.Name)
	}

	cmd := []string{
		"VBoxManage",
		"startvm",
		v.instance.Name,
		"--type",
		"gui",
	}
	_, err = v.executor.Execute(cmd)
	return
}

// Stop VM
func (v *Virtualbox) Stop() error {
	if exists, err := v.checkIfExists(); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("%s does not exists", v.instance.Name)
	}

	if isRunning, err := v.checkIfRunning(); err != nil {
		return err
	} else if !isRunning {
		return fmt.Errorf("%s is not running", v.instance.Name)
	}

	if err := v.executePreStopActions(); err != nil {
		return err
	}

	cmd := []string{
		"VBoxManage",
		"controlvm",
		v.instance.Name,
		"acpipowerbutton",
	}
	_, err := v.executor.Execute(cmd)
	return err
}

// GetState VM
func (v *Virtualbox) GetState() *State {
	return v.state
}

// CheckIsUpdated VM
func (v *Virtualbox) CheckIsUpdated() (isUpToDate bool, err error) {
	return v.updater.CheckIsUpdated()
}

// Update VM
func (v *Virtualbox) Update() (updated bool, err error) {
	isUpToDate, err := v.updater.CheckIsUpdated()
	if err != nil || isUpToDate {
		return
	}

	err = v.unregisterIfExists()
	if err != nil {
		return
	}

	err = v.updater.Update()
	if err != nil {
		return
	}

	return true, v.init()
}

func (v *Virtualbox) unregisterIfExists() (err error) {
	exists, err := v.checkIfExists()
	if err != nil {
		return
	}
	if !exists {
		return
	}
	return v.Unregister()
}

// Unregister a VM
func (v *Virtualbox) Unregister() (err error) {
	hostIf, err := v.infoFromVM("hostonlyadapter2=\"(.*)\"")
	if err != nil {
		return
	}

	rootVol, err := v.infoFromVM("SAS-0-0\"=\"(.*)\"")
	if err != nil {
		return
	}

	userVol, err := v.infoFromVM("SAS-1-0\"=\"(.*)\"")
	if err != nil {
		return
	}

	cmd0 := []string{
		"VBoxManage", "hostonlyif", "remove", hostIf,
	}
	_, err = v.executor.Execute(cmd0)
	if err != nil {
		return
	}

	cmd1 := []string{
		"VBoxManage", "storageattach", v.instance.Name,
		"--storagectl", "SAS",
		"--port", "0",
		"--medium", "none",
	}
	_, err = v.executor.Execute(cmd1)
	if err != nil {
		return
	}

	cmd2 := []string{
		"VBoxManage", "closemedium", "disk", rootVol,
	}
	_, err = v.executor.Execute(cmd2)
	if err != nil {
		return
	}

	cmd3 := []string{
		"VBoxManage", "storageattach", v.instance.Name,
		"--storagectl", "SAS",
		"--port", "1",
		"--medium", "none",
	}
	_, err = v.executor.Execute(cmd3)
	if err != nil {
		return
	}

	cmd4 := []string{
		"VBoxManage", "closemedium", "disk", userVol,
	}
	_, err = v.executor.Execute(cmd4)
	if err != nil {
		return
	}

	cmd5 := []string{
		"VBoxManage", "unregistervm", v.instance.Name, "--delete",
	}
	_, err = v.executor.Execute(cmd5)

	return
}

func (v *Virtualbox) setState(state *State) (err error) {
	oldState := v.state
	v.state = state

	if !oldState.AllSystemsReady && v.state.AllSystemsReady {
		return v.executePostStartActions()
	}

	return
}

// AddPostStartAction triggers an action just after starting the VM
func (v *Virtualbox) AddPostStartAction(f func() error) {
	v.postStartActions = append(v.postStartActions, f)
}

// AddPreStopAction triggers an action just before stopping the VM
func (v *Virtualbox) AddPreStopAction(f func() error) {
	v.postStartActions = append(v.postStartActions, f)
}

func (v *Virtualbox) executePreStopActions() error {
	for _, x := range v.preStopActions {
		if err := x(); err != nil {
			return err
		}
	}

	return nil
}

func (v *Virtualbox) executePostStartActions() error {
	for _, x := range v.postStartActions {
		if err := x(); err != nil {
			return err
		}
	}

	return nil
}

func (v *Virtualbox) init() (err error) {
	if err := v.install(); err != nil {
		return err
	}

	if err := v.setDefaultOptions(); err != nil {
		return err
	}

	if err := v.setCPU(fmt.Sprintf("%d", v.instance.Vcpu)); err != nil {
		return err
	}

	if err := v.setMEM(fmt.Sprintf("%d", v.instance.Vmem)); err != nil {
		return err
	}

	hostonlyif, err := v.createHostOnlyNetwork(v.instance.Localip)
	if err != nil {
		return err
	}

	if err := v.attachHostOnlyNetwork(hostonlyif); err != nil {
		return err
	}

	if err := v.setStorageCtl(); err != nil {
		return err
	}

	if err := v.attachRootImage(); err != nil {
		return err
	}

	fileExist, _ := util.Exists(v.userData)
	if !fileExist {
		if err := v.createUserData(fmt.Sprintf("%d", (v.userDataSize * 1024))); err != nil {
			return err
		}
	}
	if err := v.attachUserData(); err != nil {
		return err
	}

	return nil
}

func (v *Virtualbox) install() (err error) {
	cmdCreateVM := []string{
		"VBoxManage", "createvm",
		"--name", v.instance.Name,
		"--ostype", "Ubuntu_64",
		"--register",
	}

	_, err = v.executor.Execute(cmdCreateVM)

	return
}

func (v *Virtualbox) checkIfRunning() (bool, error) {
	cmd := []string{
		"VBoxManage",
		"list",
		"runningvms",
	}

	vms, err := v.executor.Execute(cmd)
	if err != nil {
		return false, err
	}

	return v.isVMPresent(vms), nil
}

func (v *Virtualbox) checkIfExists() (bool, error) {
	cmd := []string{
		"VBoxManage",
		"list",
		"vms",
	}

	vms, err := v.executor.Execute(cmd)
	if err != nil {
		return false, err
	}

	return v.isVMPresent(vms), nil
}

func (v *Virtualbox) isVMPresent(vms string) bool {
	scanner := bufio.NewScanner(strings.NewReader(vms))
	for scanner.Scan() {
		vm := strings.Fields(scanner.Text())
		if strings.Compare(v.instance.Name, strings.Trim(vm[0], "\"")) == 0 {
			return true
		}
	}

	return false
}

func (v *Virtualbox) setDefaultOptions() error {
	cmdSetOptions := []string{
		"VBoxManage", "modifyvm", v.instance.Name,
		"--acpi", "on",
		"--ioapic", "on",
		"--rtcuseutc", "on",
		"--vram", "2",
		"--accelerate3d", "off",
		"--accelerate2dvideo", "off",
		"--graphicscontroller", "VMSVGA",
		"--biosbootmenu", "disabled",
		"--bioslogofadein", "off",
		"--bioslogofadeout", "off",
		"--bioslogodisplaytime", "0",
		"--firmware", "bios",
		"--boot1", "disk",
		"--boot2", "none",
		"--boot3", "none",
		"--boot4", "none",
		"--mouse", "ps2",
		"--keyboard", "ps2",
		"--usb", "off",
		"--draganddrop", "disabled",
		"--usbcardreader", "off",
		"--audio", "none",
		"--vrde", "off",
		"--tracing-enabled", "off",
		"--nic1", "nat",
		"--nictype1", "virtio",
		"--cableconnected1", "on",
		"--nicpromisc1", "deny",
	}

	_, err := v.executor.Execute(cmdSetOptions)
	return err
}

func (v *Virtualbox) createHostOnlyNetwork(localip string) (string, error) {
	var hostonlyif string
	cmdCreate := []string{
		"VBoxManage",
		"hostonlyif",
		"create",
	}
	output, _ := v.executor.Execute(cmdCreate)
	re := regexp.MustCompile(`Interface '([^']+)' was successfully created`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		hostonlyif = matches[1]
	} else {
		return "", fmt.Errorf("could not determine the interface name from vbox output: %s", output)
	}

	localiparray := strings.Split(localip, ".")
	localip = localiparray[0] + "." + localiparray[1] + "." + localiparray[2] + ".1"
	cmdConf := []string{
		"VBoxManage", "hostonlyif",
		"ipconfig", hostonlyif,
		"--ip", localip,
		"--netmask", "255.255.255.0",
	}
	if _, err := v.executor.Execute(cmdConf); err != nil {
		return "", err
	}

	return hostonlyif, nil
}

func (v *Virtualbox) attachHostOnlyNetwork(hostonlyif string) error {
	cmd := []string{
		"VBoxManage", "modifyvm", v.instance.Name,
		"--nic2", "hostonly",
		"--nictype2", "virtio",
		"--cableconnected2", "on",
		"--nicpromisc2", "deny",
		"--hostonlyadapter2", hostonlyif,
	}
	_, err := v.executor.Execute(cmd)
	return err
}

func (v *Virtualbox) setMEM(vmem string) error {
	cmd := []string{
		"VBoxManage", "modifyvm", v.instance.Name,
		"--memory", vmem,
		"--nestedpaging", "on",
		"--largepages", "on",
		"--pae", "on",
	}
	_, err := v.executor.Execute(cmd)
	return err
}

func (v *Virtualbox) setCPU(vcpu string) error {
	cmd := []string{
		"VBoxManage", "modifyvm", v.instance.Name,
		"--cpus", vcpu,
		"--cpu-profile", "host",
		"--hwvirtex", "on",
		"--paravirtprovider", "kvm",
		"--vtxvpid", "on",
		"--vtxux", "on",
	}
	_, err := v.executor.Execute(cmd)
	return err
}

func (v *Virtualbox) setStorageCtl() error {
	cmd := []string{
		"VBoxManage", "storagectl", v.instance.Name,
		"--name", "SAS",
		"--add", "sas",
		"--controller", "LSILogicSAS",
		"--portcount", "2",
		"--hostiocache", "on",
		"--bootable", "on",
	}
	_, err := v.executor.Execute(cmd)
	return err
}

func (v *Virtualbox) attachRootImage() error {
	cmd := []string{
		"VBoxManage", "storageattach", v.instance.Name,
		"--storagectl", "SAS",
		"--port", "0",
		"--type", "hdd",
		"--medium", v.boxPath,
		"--mtype", "normal",
		"--nonrotational", "on",
		"--discard", "on",
	}
	_, err := v.executor.Execute(cmd)
	return err
}

func (v *Virtualbox) createUserData(size string) error {
	cmd := []string{
		"VBoxManage", "createmedium", "disk",
		"--filename", v.userData,
		"--size", size,
		"--format", "VDI",
		"--variant", "Standard",
	}
	_, err := v.executor.Execute(cmd)
	return err
}

func (v *Virtualbox) attachUserData() error {
	cmd := []string{
		"VBoxManage", "storageattach", v.instance.Name,
		"--storagectl", "SAS",
		"--port", "1",
		"--type", "hdd",
		"--medium", v.userData,
		"--mtype", "normal",
		"--nonrotational", "on",
		"--discard", "on",
	}
	_, err := v.executor.Execute(cmd)
	return err
}

func (v *Virtualbox) infoFromVM(search string) (string, error) {
	cmd := []string{
		"VBoxManage", "showvminfo",
		"--machinereadable", v.instance.Name,
	}
	stdOut, err := v.executor.Execute(cmd)

	re := regexp.MustCompile(fmt.Sprintf(`%s`, search))
	matches := re.FindStringSubmatch(stdOut)
	if len(matches) > 0 {
		return matches[1], err
	}

	return "", fmt.Errorf("unable to retrieve information from the Virtual Machine")
}
