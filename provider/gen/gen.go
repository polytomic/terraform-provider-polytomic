package main

import (
	"log"
	"os"

	"github.com/polytomic/terraform-provider-polytomic/provider/gen/connections"
)

func main() {
	var err error
	if len(os.Args) > 0 && os.Args[len(os.Args)-1] == "--sort" {
		err = connections.SortConnections()
	} else {
		err = connections.GenerateConnections()
	}
	if err != nil {
		log.Fatal(err.Error())
	}
}
