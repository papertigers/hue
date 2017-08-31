package main

import (
	"fmt"
	"log"

	"github.com/papertigers/hue/lib/bridge"
)

func main() {
	timeout := 30
	fmt.Printf("Discovering bridges.  Please wait (%ds)....\n\n", timeout)
	bridges, err := bridge.Discover(timeout)
	if err != nil {
		log.Fatalln(err)
	}
	for bridge, _ := range bridges {
		fmt.Printf("%v\n", bridge)
	}
	//bridge := &bridge.Bridge{
	//	IP: "10.0.1.80",
	//}
	//
	//	res, err := bridge.CreateUser()
	//	if err != nil {
	//		log.Fatalln(err)
	//	}
	//	fmt.Printf("%v\n", res)
}
