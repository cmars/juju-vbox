package main

import (
	"fmt"
	"log"
	"os"

	"github.com/riobard/go-virtualbox"
)

func die(err error) {
	log.Fatalln(err)
}

func main() {
	if len(os.Args) < 2 {
		die(fmt.Errorf("Usage: <img file>"))
	}
	imgfile := os.Args[1]
	m, err := virtualbox.CreateMachine("foo", "")
	if err != nil {
		die(err)
	}
	err = m.AddStorageCtl("storecntl", virtualbox.StorageController{
		SysBus:      virtualbox.SysBusSATA,
		Ports:       1,
		Chipset:     virtualbox.CtrlIntelAHCI,
		HostIOCache: true,
		Bootable:    true,
	})
	if err != nil {
		die(err)
	}
	err = m.AttachStorage("storecntl", virtualbox.StorageMedium{
		DriveType: virtualbox.DriveHDD,
		Medium:    imgfile,
	})
	fmt.Println(m)
}
