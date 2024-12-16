package main

import (
	"context"
	"log"

	"github.com/polytomic/terraform-provider-polytomic/provider/gen/connections"
)

func main() {
	err := connections.GenerateConnections(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}
}
