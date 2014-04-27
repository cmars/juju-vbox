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

const (
	ifac0Name = "vboxnet0"
)

func CreateFirstIFac(m *vbox.Machine) (err error) {
	ifacName := ifac0Name
	honets, err := vbox.HostonlyNets()
	if err != nil {
		return
	}
	hon, ok := honets[ifacName]
	if !ok {
		hon, err = vbox.CreateHostonlyNet()
		if err != nil {
			return
		}
		ifacName = hon.Name
		hon.IPv4 = net.IPNet{
			IP:   net.IPv4(192, 168, 50, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		}
		err = hon.Config()
		if err != nil {
			return
		}
	}
	err = m.SetNIC(2, vbox.NIC{
		Network:         vbox.NICNetHostonly,
		Hardware:        vbox.IntelPro1000MTDesktop,
		HostonlyAdapter: hon.Name,
	})
	if err != nil {
		return
	}
	dhcps, err := vbox.DHCPs()
	if err != nil {
		return
	}
	log.Println("DHCP Servers", dhcps)
	dhcp, ok := dhcps[ifacName]
	if !ok {
		dhcp = &vbox.DHCP{
			NetworkName: ifacName,
			IPv4: net.IPNet{
				IP:   net.IPv4(192, 168, 50, 1),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
			LowerIP: net.IPv4(192, 168, 50, 2),
			UpperIP: net.IPv4(192, 168, 50, 250),
		}
	}
	dhcp.Enabled = true // Make sure it's turned on.
	vbox.AddHostonlyDHCP(ifacName, *dhcp)
	return
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
	err = CreateFirstIFac(m)
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
