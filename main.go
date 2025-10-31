package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"libvirt.org/go/libvirt"

	"stuvps.app/vm-ctrl-api/models"
	"stuvps.app/vm-ctrl-api/vm"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := libvirt.NewConnect(os.Getenv("QEMU_URL"))
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
			InterfaceName: req.Resources.NetworkInterface,
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

	r.GET("/console",  func(c *gin.Context) {
		
	})
	r.Run()
}
