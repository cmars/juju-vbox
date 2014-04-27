package main

import (
	"fmt"
	"log"
	"net"
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
	honets, err := vbox.HostonlyNets()
	if err != nil {
		return nil, err
	}
	hon, ok := honets["vboxnet0"]
	if !ok {
		hon, err = vbox.CreateHostonlyNet()
		if err != nil {
			return nil, err
		}
		hon.Name = "vboxnet0"
		hon.IPv4 = net.IPNet{
			IP:   net.IPv4(192, 168, 50, 1),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		}
		//hon.DHCP = true
		err = hon.Config()
		if err != nil {
			return nil, err
		}
	}
	err = m.SetNIC(1, vbox.NIC{
		Network:         vbox.NICNetHostonly,
		Hardware:        vbox.IntelPro1000MTDesktop,
		HostonlyAdapter: hon.Name,
	})
	if err != nil {
		return nil, err
	}
	err = m.SetNIC(2, vbox.NIC{
		Network:  vbox.NICNetNAT,
		Hardware: vbox.IntelPro1000MTDesktop,
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func main() {
	vbox.Verbose = true
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
	/*
		err = m.AddNATPF(1, "ssh", vbox.PFRule{
			Proto:     vbox.PFTCP,
			HostIP:    nil,
			HostPort:  2222,
			GuestIP:   nil,
			GuestPort: 22,
		})
		if err != nil {
			die(err)
		}
	*/
	fmt.Println(m)
}
