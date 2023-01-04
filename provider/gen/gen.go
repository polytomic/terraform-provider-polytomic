package main

import (
	"log"

	"github.com/polytomic/terraform-provider-polytomic/provider/gen/connections"
)

func main() {
	err := connections.GenerateConnections() // Generate connections
	if err != nil {
		log.Fatal(err.Error())
	}
}
