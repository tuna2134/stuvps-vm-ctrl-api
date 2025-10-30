package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"libvirt.org/go/libvirt"
	"stuvps.app/vm-ctrl-api/models"
	"stuvps.app/vm-ctrl-api/vm"
)

func main() {
	println("Hello, World!")
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/vms", func(c *gin.Context) {
		var req models.PostVMRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		vmName := uuid.New().String()

		config := vm.VMConfig{
			BasePath:      "/var/lib/libvirt/images/",
			Name:          vmName,
			Memory:        req.Resources.Memory,
			VCPUs:         req.Resources.VCPUs,
			InterfaceName: "virbr0",
			Password:      req.CloudInit.Password,
			Network: vm.VMConfigNetwork{
				IPAddress: req.CloudInit.IPAddress,
				Gateway:   req.CloudInit.Gateway,
			},
			Script: req.CloudInit.Script,
		}

		if err := vm.CreateVM(conn, config); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "VM created successfully", "vm_name": vmName})
	})
	r.Run()
}
