package cloud_init

import (
	"os"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/goccy/go-yaml"
)

type CloudConfig struct {
	Hostname   string      `yaml:"hostname"`
	SSHPWAuth  bool        `yaml:"ssh_pwauth"`
	Users      []UserData  `yaml:"users"`
	RunCMD     []string    `yaml:"runcmd,omitempty"`
	WriteFiles []WriteFile `yaml:"write_files,omitempty"`
}

type WriteFile struct {
	Content     string `yaml:"content"`
	Path        string `yaml:"path"`
	Owner       string `yaml:"owner,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
}

type UserData struct {
	Name            string `yaml:"name"`
	PlainTextPasswd string `yaml:"plain_text_passwd"`
	LockPasswd      bool   `yaml:"lock_passwd"`
	Shell           string `yaml:"shell"`
	Sudo            string `yaml:"sudo"`
}

type NetworkConfig struct {
	Network NetworkConfigNetwork `yaml:"network"`
}

type NetworkConfigNetwork struct {
	Ethernets map[string]NetworkConfigEthernet `yaml:"ethernets"`
	Version   int                              `yaml:"version"`
}

type NetworkConfigEthernet struct {
	Match       NetworkConfigEthernetMatch   `yaml:"match"`
	DHCP4       bool                         `yaml:"dhcp4"`
	Addresses   []string                     `yaml:"addresses,omitempty"`
	Routes      []NetworkConfigEthernetRoute `yaml:"routes,omitempty"`
	Nameservers NetworkConfigNameservers     `yaml:"nameservers"`
}

type NetworkConfigNameservers struct {
	Addresses []string `yaml:"addresses,omitempty"`
}

type NetworkConfigEthernetRoute struct {
	To  string `yaml:"to"`
	Via string `yaml:"via"`
}

type NetworkConfigEthernetMatch struct {
	MAC string `yaml:"macaddress"`
}

func GenerateCloudConfig(password string, script string) ([]byte, error) {
	data := new(CloudConfig)
	data.Hostname = "localhost"
	data.SSHPWAuth = true
	data.RunCMD = []string{"bash", "/root/setup_script.sh"}
	data.Users = []UserData{
		{
			Name:            "ubuntu",
			PlainTextPasswd: password,
			LockPasswd:      false,
			Shell:           "/bin/bash",
			Sudo:            "ALL=(ALL) NOPASSWD:ALL",
		},
	}
	data.WriteFiles = []WriteFile{
		{
			Content:     script,
			Path:        "/root/setup_script.sh",
			Owner:       "root:root",
			Permissions: "0755",
		},
	}
	b, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}
	b = append([]byte("#cloud-config\n"), b...)
	return b, nil
}

func GenerateNetworkConfig(mac string, address string, gateway string) ([]byte, error) {
	data := new(NetworkConfig)
	data.Network.Ethernets = make(map[string]NetworkConfigEthernet)
	data.Network.Version = 2
	data.Network.Ethernets["eth0"] = NetworkConfigEthernet{
		Match: NetworkConfigEthernetMatch{
			MAC: mac,
		},
		DHCP4:     false,
		Addresses: []string{address},
		Routes: []NetworkConfigEthernetRoute{
			{
				To:  "default",
				Via: gateway,
			},
		},
		Nameservers: NetworkConfigNameservers{
			Addresses: []string{"1.1.1.1"},
		},
	}
	b, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}
	b = append([]byte("#cloud-config\n"), b...)
	return b, nil
}

func unused(_ ...any) {
}

func CreateDisk(
	path string, password string, mac string,
	address string, gateway string, script string,
) error {
	// cloud-initで扱うISOイメージを作成します。
	// 10MBのサイズで作成します。
	d, err := diskfs.Create(path, 10*1024*1024, diskfs.SectorSizeDefault)
	if err != nil {
		return err
	}
	d.LogicalBlocksize = 2048
	fs, err := d.CreateFilesystem(disk.FilesystemSpec{
		FSType:      filesystem.TypeISO9660,
		VolumeLabel: "CIDATA",
	})
	if err != nil {
		return err
	}
	// user-dataを作成します。
	// ここでは、パスワード認証を有効にしたユーザーを作成します。
	// ユーザー名は"ubuntu"、パスワードは引数で指定されたものを使用します。
	// シェルは"/bin/bash"、sudo権限は"ALL=(ALL) NOPASSWD:ALL"を設定します。
	rw, err := fs.OpenFile("user-data", os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	b, err := GenerateCloudConfig(password, script)
	if err != nil {
		return err
	}
	written, err := rw.Write(b)
	if err != nil {
		return err
	}
	unused(written)
	rw.Close()
	// meta-dataを作成します。
	// ここでは、仮のインスタンスIDとホスト名を設定します。
	rw, err = fs.OpenFile("meta-data", os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	metaData := "instance-id: i-1234567890\nlocal-hostname: localhost\n"
	written, err = rw.Write([]byte(metaData))
	if err != nil {
		return err
	}
	unused(written)
	rw.Close()
	// network-configを作成します。
	// ここでは、MACアドレス、IPアドレス、ゲートウェイを設定します。
	// MACアドレスは引数で指定されたものを使用します。
	// IPアドレスとゲートウェイも引数で指定されたものを使用します。
	networkConfig, err := GenerateNetworkConfig(
		mac, address, gateway,
	)
	if err != nil {
		return err
	}
	rw, err = fs.OpenFile("network-config", os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	written, err = rw.Write(networkConfig)
	if err != nil {
		return err
	}
	unused(written)
	rw.Close()
	// ISO9660ファイルシステムを最終化します。
	// ここでは、ボリューム識別子を"CIDATA"に設定し、
	// Rock Ridge拡張を有効にします。
	// Rock Ridge拡張は、UNIX系のファイルシステムで
	// より多くの情報を提供するために使用されます。
	// これにより、ファイルのパーミッションや所有者情報などが正しく保存されます。
	iso, ok := fs.(*iso9660.FileSystem)
	if !ok {
		return err
	}
	err = iso.Finalize(iso9660.FinalizeOptions{
		VolumeIdentifier: "CIDATA",
		RockRidge:        true,
	})
	if err != nil {
		return err
	}
	defer iso.Close()
	return nil
}
