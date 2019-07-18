package main

import (
	"wallet/client"
)

func main() {
	cli := client.NewCLI("./data", "http://localhost:8545")

	cli.Run()
}
