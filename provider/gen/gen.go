package main

import (
	"log"

	"github.com/polytomic/terraform-provider-polytomic/provider/gen/connections"
)

func main() {
	err := connections.GenerateConnections()
	if err != nil {
		log.Fatal(err.Error())
	}
}
