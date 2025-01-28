package main

import (
	"fmt"
	"github.com/Noah-Wilderom/cfx"
	"syscall/js"
)

var (
	c  = make(chan bool)
	fn = js.FuncOf
)

func init() {
	cfx.Print(fmt.Sprintf("wasm %s loaded.", cfx.Server.GetCurrentResourceName()))
}

func main() {
	cfx.Server.RegisterCommand("test-go", fn(func(this js.Value, args []js.Value) interface{} {
		cfx.TriggerClientEvent(
			"chat:addMessage",
			-1,
			"Test message",
		)
		cfx.Print("Test print")

		return nil
	}), false)

	// <-c
}
