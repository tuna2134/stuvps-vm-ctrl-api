package gen

import (
	"libvirt.org/libvirt-go-xml"
	"crypto/rand"
	"net"
)

func GenerateMACAddress() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	buf[0] &= 0b11111110
	buf[0] |= 0b00000010
	return net.HardwareAddr(buf).String()
}

func CreateDomainXML(basePath string, name string, memory int, vcpus int, mac string, interfaceName string) (string, error) {
	qcow2Path := basePath + "/" + name + "/image.qcow2"
	domain := libvirtxml.Domain{
		Type: "kvm",
		Name: name,
		Memory: &libvirtxml.DomainMemory{
			Value: uint(memory),
			Unit:  "GB",
		},
		Clock: &libvirtxml.DomainClock{
			Offset: "utc",
		},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "destroy",
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "pc",
				Type:    "hvm",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{
					Dev: "hd",
				},
			},
			ACPI: &libvirtxml.DomainACPI{},
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: uint(vcpus),
		},
		CPU: &libvirtxml.DomainCPU{
			Mode: "host-passthrough",
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
			PAE:  &libvirtxml.DomainFeature{},
			PMU:  &libvirtxml.DomainFeatureState{},
		},
		Devices: &libvirtxml.DomainDeviceList{
			Emulator: "/usr/bin/qemu-system-x86_64",
			Disks: []libvirtxml.DomainDisk{
				{
					Driver: &libvirtxml.DomainDiskDriver{
						Name:  "qemu",
						Type:  "qcow2",
						Cache: "writeback",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: qcow2Path,
						},
						Index: 0,
					},
					BackingStore: &libvirtxml.DomainDiskBackingStore{},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vda",
						Bus: "virtio",
					},
					Alias: &libvirtxml.DomainAlias{
						Name: "virtio-disk0",
					},
					Address: &libvirtxml.DomainAddress{
						PCI: &libvirtxml.DomainAddressPCI{
							Domain:   &[]uint{0x0000}[0],
							Bus:      &[]uint{0x04}[0],
							Slot:     &[]uint{0x01}[0],
							Function: &[]uint{0x0}[0],
						},
					},
				},
				{
					Driver: &libvirtxml.DomainDiskDriver{
						Type: "raw",
						Name: "qemu",
					},
					Device: "cdrom",
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "hdc",
						Bus: "ide",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: basePath + name + "/seek.iso",
						},
					},
					ReadOnly: &libvirtxml.DomainDiskReadOnly{},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Source: &libvirtxml.DomainInterfaceSource{
						Bridge: &libvirtxml.DomainInterfaceSourceBridge{
							Bridge: interfaceName,
						},
					},
					MAC: &libvirtxml.DomainInterfaceMAC{
						Address: mac,
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
				},
			},
			Graphics: []libvirtxml.DomainGraphic{
				{
					VNC: &libvirtxml.DomainGraphicVNC{
						Port:   -1,
						Listen: "0.0.0.0",
					},
				},
			},
			Serials: []libvirtxml.DomainSerial{
				{
					Protocol: &libvirtxml.DomainChardevProtocol{
						Type: "pty",
					},
					Target: &libvirtxml.DomainSerialTarget{
						Port: &[]uint{0}[0],
					},
					Source: &libvirtxml.DomainChardevSource{
						Pty: &libvirtxml.DomainChardevSourcePty{
							Path: "/dev/pts/0",
						},
					},
				},
			},
		},
	}
	return domain.Marshal()
}