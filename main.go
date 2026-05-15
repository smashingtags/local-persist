package main

import (
	"log"

	"github.com/docker/go-plugins-helpers/volume"
)

func main() {
	driver := newLocalPersistDriver()
	handler := volume.NewHandler(driver)
	log.Fatal(handler.ServeUnix("root", driver.name))
}
