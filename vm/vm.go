package vm

import (
	"stuvps.app/vm-ctrl-api/cloud_init"
	"stuvps.app/vm-ctrl-api/vm/gen"
	"libvirt.org/go/libvirt"
)

type VMConfig struct {
	BasePath	  string
	Name		  string
	Memory		  int
	VCPUs		  int
	InterfaceName string
	Password	  string
	Network		  VMConfigNetwork
	Script    	  string
}

type VMConfigNetwork struct {
	IPAddress string
	Gateway   string
}

func CreateVM(conn *libvirt.Connect, config VMConfig) error {
	mac := gen.GenerateMACAddress()
	cloud_init.CreateDisk(
		config.BasePath + config.Name + "/seek.iso",
		config.Password,
		mac,
		config.Network.IPAddress,
		config.Network.Gateway,
		config.Script,
	)
	domainXML, err := gen.CreateDomainXML(config.BasePath, config.Name, config.Memory, config.VCPUs, mac, config.InterfaceName)
	if err != nil {
		return err
	}

	domain, err := conn.DomainDefineXML(domainXML)
	if err != nil {
		return err
	}
	defer domain.Free()

	if err := domain.Create(); err != nil {
		return err
	}

	return nil
}