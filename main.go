package main

import (
	"fmt"
	"log"

	"github.com/papertigers/hue/lib/bridge"
)

func main() {
	fmt.Printf("Discovering bridges.  Please wait (10s)....\n\n")
	bridges, err := bridge.Discover()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%v\n", bridges)
}
