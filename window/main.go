package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/inancgumus/screen"
)

func main() {
	fmt.Print("testing")
	time.Sleep(time.Second)
	fmt.Print("\rhi hi hi\n")
	fmt.Println("cool")
	time.Sleep(time.Second)
	fmt.Print("\r???\n")
	cmd := exec.Command("clear")
	err := cmd.Run()
	if err != nil {
		fmt.Println("err:%v\n", err)
	}
	time.Sleep(time.Second)
	fmt.Println("hello")
	time.Sleep(time.Second)
	screen.Clear()
	screen.MoveTopLeft()
	fmt.Println("hello")

}
