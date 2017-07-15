package main

import "fmt"

func main() {
	c, err := NewConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	swallow := NewSwallow(c)
	swallow.ShutdownHook()
	swallow.Run()
}
