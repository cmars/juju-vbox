package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"

	"code.google.com/p/go-uuid/uuid"
	vbox "github.com/riobard/go-virtualbox"
)

func die(err error) {
	log.Fatalln(err)
}

const (
	storageCtrlName = "storctrl0"
	ifac0Name       = "vboxnet0"
)

func CreateFirstIFac(m *vbox.Machine) error {
	ifacName := ifac0Name // TODO: configurable?
	ifacKey := fmt.Sprintf("HostInterfaceNetworking-%s", ifacName)
	honets, err := vbox.HostonlyNets()
	if err != nil {
		return err
	}
	log.Println("Host Only Nets:", honets)
	hon, ok := honets[ifacKey]
	if !ok {
		hon, err = vbox.CreateHostonlyNet()
		if err != nil {
			return err
		}
		ifacName = hon.Name
		hon.IPv4 = net.IPNet{
			IP:   net.IPv4(192, 168, 50, 1),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		}
		err = hon.Config()
		if err != nil {
			return err
		}
	} else {
		err = hon.Config()
		if err != nil {
			return err
		}
		err = m.SetNIC(1, vbox.NIC{
			Network:         vbox.NICNetHostonly,
			Hardware:        vbox.IntelPro1000MTDesktop,
			HostonlyAdapter: hon.Name,
		})
		if err != nil {
			return err
		}
	}

	dhcps, err := vbox.DHCPs()
	if err != nil {
		return err
	}
	log.Println("DHCP Servers", dhcps)
	dhcp, ok := dhcps[ifacKey]
	if !ok {
		dhcp = &vbox.DHCP{
			NetworkName: ifacName,
			IPv4: net.IPNet{
				IP:   hon.IPv4.IP,
				Mask: hon.IPv4.Mask,
			},
			LowerIP: net.IPv4(192, 168, 50, 2),
			UpperIP: net.IPv4(192, 168, 50, 250),
		}
		dhcp.Enabled = true // Make sure it's turned on.
		err = vbox.AddHostonlyDHCP(ifacName, *dhcp)
	}
	return err
}

func CreateMachineWithImage(name, imgfile string) (*vbox.Machine, error) {
	// Create a hard link to the image file. Virtualbox likes to delete the
	// image with the machine.
	imglink := path.Join(path.Dir(imgfile),
		fmt.Sprintf(".%s.%s", path.Base(imgfile), uuid.New()))
	err := os.Link(imgfile, imglink)
	if err != nil {
		return nil, err
	}

	m, err := vbox.CreateMachine(name, "")
	if err != nil {
		return nil, err
	}
	err = m.AddStorageCtl(storageCtrlName, vbox.StorageController{
		SysBus:      vbox.SysBusSATA,
		Ports:       1,
		Chipset:     vbox.CtrlIntelAHCI,
		HostIOCache: true,
		Bootable:    true,
	})
	if err != nil {
		return nil, err
	}
	err = m.AttachStorage(storageCtrlName, vbox.StorageMedium{
		DriveType: vbox.DriveHDD,
		Medium:    imglink,
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
	err = m.Start()
	if err != nil {
		die(err)
	}
	for {
		err = m.Refresh()
		if err != nil {
			fmt.Println("refresh:", err)
		} else {
			fmt.Println(m)
		}
		time.Sleep(5 * time.Second)
	}
}
