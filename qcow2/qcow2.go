package qcow2

import (
	"os/exec"
	"strconv"
)

func ConvertImage(inputPath, outputPath string) error {
	cmd := exec.Command("qemu-img", "convert", "-O", "qcow2", inputPath, outputPath)
	return cmd.Run()
}

func ResizeImage(path string, size int) error {
	sizeMB := size * 1024
	cmd := exec.Command("qemu-img", "resize", path, strconv.Itoa(sizeMB)+"MB")
	return cmd.Run()
}
