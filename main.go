package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	//vbox "github.com/riobard/go-virtualbox"
	vbox "github.com/gdey/go-virtualbox"
)

func die(err error) {
	log.Fatalln(err)
}

const (
	ifac0Name = "vboxnet0"
)

var (
	ErrDHCPNotExist        = errors.New("DHCP setting does not exist.")
	ErrHostonlyNetNotExist = errors.New("Hostonly network does not exist.")
)

func getDHCPForIPNet(ipNet net.IPNet) (*vbox.DHCP, error) {
	dhcps, err := vbox.DHCPs()
	if err != nil {
		return nil, err
	}
	log.Println("DHCP Servers", dhcps)
	for _, dhcp := range dhcps {
		if ipNet.String() == dhcp.IPv4.String() {
			return dhcp, nil
		}
	}
	return nil, ErrDHCPNotExist
}
func getHostonlyForIPNet(ipNet net.IPNet) (*vbox.HostonlyNet, error) {
	nets, err := vbox.HostonlyNets()
	if err != nil {
		return nil, err
	}
	for _, honet := range nets {
		if ipNet.String() == honet.IPv4.String() {
			return honet, nil
		}
	}
	return nil, ErrHostonlyNetNotExist
}

func findCreateHostonlyForIPNet(ipNet net.IPNet) (*vbox.HostonlyNet, error) {
	hon, err := getHostonlyForIPNet(ipNet)
	if err == ErrHostonlyNetNotExist {
		hon, err := vbox.CreateHostonlyNet()
		if err != nil {
			return nil, err
		}
		hon.IPv4 = ipNet
		err = hon.Config()
		if err != nil {
			return nil, err
		}
		dhcp, err := getDHCPForIPNet(ipNet)
		if err == ErrDHCPNotExist {
			dhcp = &vbox.DHCP{
				NetworkName: hon.Name,
				IPv4:        ipNet,
				LowerIP:     net.IPv4(172, 16, 16, 2),
				UpperIP:     net.IPv4(172, 16, 16, 240),
			}
		}
		dhcp.Enabled = true // Make sure it's turned on.
		vbox.AddHostonlyDHCP(hon.Name, *dhcp)
		return hon, nil
	}
	return hon, nil
}

func CreateFirstIFac(m *vbox.Machine, nicNumber int) (err error) {
	hon, err := findCreateHostonlyForIPNet(net.IPNet{
		IP:   net.IPv4(172, 16, 16, 1),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	})

	err = m.SetNIC(nicNumber, vbox.NIC{
		Network:         vbox.NICNetHostonly,
		Hardware:        vbox.IntelPro1000MTDesktop,
		HostonlyAdapter: hon.Name,
	})
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
	//TODO: We should clone the imagefile first, and the
	//      Attach it — this will make the system take a bit
	//      longer, but this is the only way to boot multiple
	//      machines.
	vbox.CloneDiskImage(name+".vmdk", imgfile)
	err = m.AttachStorage("storecntl", vbox.StorageMedium{
		DriveType: vbox.DriveHDD,
		Medium:    name + ".vmdk",
	})
	if err != nil {
		return nil, err
	}
	/*
		err = m.SetNIC(1, vbox.NIC{
			Network:  vbox.NICNetNAT,
			Hardware: vbox.IntelPro1000MTDesktop,
		})
		if err != nil {
			return nil, err
		}
	*/
	err = CreateFirstIFac(m, 1)
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
