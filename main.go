package main

import (
	"fmt"
	"log"
	"os"

	vbox "github.com/riobard/go-virtualbox"
)

func die(err error) {
	log.Fatalln(err)
}

func CreateMachineWithImage(name, imgfile string) (*vbox.Machine, error) {
	m, err := vbox.CreateMachine(name, "")
	if err != nil {
		return nil, err
	}
	err = m.AddStorageCtl("storecntl", vbox.StorageController{
		SysBus:      vbox.SysBusSATA,
		Ports:       1,
		Chipset:     vbox.CtrlIntelAHCI,
		HostIOCache: true,
		Bootable:    true,
	})
	if err != nil {
		return nil, err
	}
	err = m.AttachStorage("storecntl", vbox.StorageMedium{
		DriveType: vbox.DriveHDD,
		Medium:    imgfile,
	})
	if err != nil {
		return nil, err
	}
	err = m.SetNIC(1, vbox.NIC{
		Network:  vbox.NICNetNAT,
		Hardware: vbox.IntelPro1000MTDesktop,
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func main() {
	//vbox.Verbose = true
	if len(os.Args) < 3 {
		die(fmt.Errorf("Usage: <name> <imgfile>"))
	}
	name := os.Args[1]
	imgfile := os.Args[2]
	m, err := vbox.GetMachine(name)
	if err == vbox.ErrMachineNotExist {
		m, err = CreateMachineWithImage(name, imgfile)
	}
	if err != nil {
		die(err)
	}
	fmt.Println(m)
}
