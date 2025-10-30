package models;

type PostVMRequest struct {
	CloudInit  PostVMRequestCloudInit `json:"cloud_init"`
	Resources  PostVMRequestResources `json:"resources"`
}

type PostVMRequestCloudInit struct {
	Script	   string `json:"script"`
	Password   string `json:"password"`
	IPAddress  string `json:"ip_address"`
	Gateway    string `json:"gateway"`
}

type PostVMRequestResources struct {
	Memory int `json:"memory"`
	VCPUs  int `json:"vcpus"`
	Disk   int `json:"disk"`
}