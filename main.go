package main

import (
	"fmt"
	"gopkg.in/ukautz/clif.v1"
	"os/exec"
	"bytes"
	"strings"
)

func unbindVfio(device string)  (string) {
	//cmd :=	"sh -c \"echo 0000:01:00.0 > /sys/bus/pci/drivers/vfio-pci/unbind\""
	cmd :=	"sh -c \"echo " + device + " > /sys/bus/pci/drivers/vfio-pci/unbind\""
	fmt.Println("$", cmd)

	err, out, stderr := Shellout(cmd)

	if err != nil {
		fmt.Println("err", err)
		fmt.Println("stderr", stderr)
	}

	return out
}

func bindVfio(device string) (string){
	cmd :=	"sh -c \"echo 0000:01:00.0 > /sys/bus/pci/drivers/vfio-pci/bind\""
	//cmd :=	"sh -c \"echo " + device + " > /sys/bus/pci/drivers/vfio-pci/bind\""
	fmt.Println("$", cmd)

	err, out, stderr := Shellout(cmd)

	if err != nil {
		fmt.Println("err", err)
		fmt.Println("stderr", stderr)
	}

	return out
}

func loadNvidia() (string) {
	cmd := "modprobe nvidia"
	fmt.Println("$", cmd)

	err, out, stderr := Shellout(cmd)

	if err != nil {
		fmt.Println("err", err)
		fmt.Println("stderr", stderr)
	}

	return out
}

func unloadNvidia() (string){
	modules := lsmod()
	//fmt.Println("modukes", modules)
	var output string

	for _, module := range modules {
		cmd := "rmmod " + module
		fmt.Println("$", cmd)

		err, out, stderr := Shellout(cmd)
		output += "\n"+out

		if err != nil {
			fmt.Println("err", err)
			fmt.Println("stderr", stderr)
		}
	}
	return output
}

func lsmod() ([]string) {
	cmd := "lsmod | grep -e nvidia"
	fmt.Println("$", cmd)

	err, out, stderr := Shellout(cmd)

	if err != nil {
		fmt.Println("err", err)
		fmt.Println("stderr", stderr)
	}

	items := strings.Fields(out)

	var modules []string

	for _, item := range items {
		if strings.Contains(item, "nvidia") {
			if !contains(modules, "nvidia") {
				modules = append(modules, item)
			}
		}
	}

	return modules
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}


func Shellout(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

func status(c *clif.Command) (string){ //Arg ~ options
	device := c.Option("device").String()

	arg := "lspci -nnk -s \"" + device + "\""
	fmt.Println("$", arg)

	err, out, stderr := Shellout(arg)

	if err != nil {
		fmt.Println("err", err)
		fmt.Println("stderr", stderr)
	}

	var usedDriver string
	if strings.Contains(out, "Kernel driver in use") {
		usedDriver := strings.Fields(strings.Split(out, "Kernel driver in use:")[1])[0]

		fmt.Println("You're using \""+usedDriver+"\" as driver")
	} else {
		usedDriver = ""
	}


	return usedDriver
}

func bind(c *clif.Command) { //Arg ~ options
	device := c.Option("device").String()
	target := c.Option("target").String()

	used := status(c)
	if used == "nvidia" {

		if target == "nvidia" {
			return
		}

		if target == "vfio" || target == "vfio-pci"  || target == ""{
			fmt.Println(unloadNvidia())
			fmt.Println(bindVfio(device))
		}

	} else if used == "vfio-pci" {
		if target == "vfio" || target == "vfio-pci" {
			return
		}

		if target == "nvidia" || target == "" {
			fmt.Println(unbindVfio(device))
			fmt.Println(loadNvidia())
		}

	} else if used == "" {
		if target  == "vfio" || target == "vfio-pci" {
			fmt.Println(bindVfio(device))
		} else if target == "nvidia" {
			fmt.Println(loadNvidia())
		}
	}
}

func unbind(c *clif.Command) { //Arg ~ options
	device := c.Option("device").String()
	target := c.Option("target").String()

	if target == "nvidia" {
		fmt.Println(unloadNvidia())
		return
	}

	if target == "vfio" || target == "vfio-pci" {
		fmt.Println(unbindVfio("0000:"+device))
		return
	}
}

func main()  {
	fmt.Println("I need root!")

	cli := clif.New("bumblebee-vfio", "0.0.1", "Switch between NVIDIA drivers and vfio-pci \n for using your 3D-card within a VM.")

	device := clif.NewOption("device", "d", "Device PCI-Id. Format: -d 01:00.0", "01:00.0", true, false)
	target := clif.NewOption("target", "t", "Target driver to be used (nvidia/vfio)\n if target is empty it will be toggled", "", false, false)

	status := clif.NewCommand("status", "Return status", status)
	bind := clif.NewCommand("bind", "Bind/load kernel module", bind)
	unbind := clif.NewCommand("unbind", "Unbind/-load kernel module", unbind)

	bind.AddOption(target)
	unbind.AddOption(target)

	cli.AddDefaultOptions(device)

	cli.Add(status)
	cli.Add(bind)
	cli.Add(unbind)

	cli.Run()
}
