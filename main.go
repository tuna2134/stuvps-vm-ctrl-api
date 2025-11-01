package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"libvirt.org/go/libvirt"

	"stuvps.app/vm-ctrl-api/models"
	"stuvps.app/vm-ctrl-api/vm"
)

var wsupgrader = websocket.Upgrader{
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

	r.GET("/vms/:vmId/console/", func(c *gin.Context) {
		domain, err := conn.LookupDomainByName(c.Param("vmId"))
		if err != nil {
			c.JSON(404, gin.H{"error": "VM not found"})
			return
		}
		defer domain.Free()
		wsConn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to set websocket upgrade: %+v\n", err)
			return
		}
		defer wsConn.Close()

		stream, err := conn.NewStream(1)
		err = domain.OpenConsole("serial0", stream, 2)
		if err != nil {
			log.Printf("Failed to set websocket upgrade: %+v\n", err)
			return
		}
		defer stream.Free()

		ConsoleReciever := func() {
			for {
				p := make([]byte, 1024)
				size, err := stream.Recv(p)
				if err != nil {
					log.Fatalf("ConsoleReceiverError: %v", err)
					return
				}
				// tune the byte size
			}
		}
		go ConsoleReciever()

		WebsocketReciever := func() {
			var msg models.ConsoleMessage
			err := wsConn.ReadJSON(&msg)
			if err != nil {
				log.Printf("Failed to read websocket message: %+v\n", err)
			}
		}
		WebsocketReciever()
	})
	r.Run()
}
