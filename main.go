package main

import (
	"github.com/gin-gonic/gin"
	"libvirt.org/go/libvirt"
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
	r.Run()
}