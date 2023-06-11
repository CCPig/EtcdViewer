package main

import (
	"log"
	"os"
)

func main() {
	err := os.Setenv("FYNE_FONT", "msyhl.ttc")
	if err != nil {
		log.Fatalf("%v\n", err)
		return
	}
	EtcdView()
}
