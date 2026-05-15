package main

import (
	"log"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/smashingtags/local-persist/internal/driver"
)

func main() {
	d := driver.New()
	handler := volume.NewHandler(d)
	log.Fatal(handler.ServeUnix("root", "local-persist"))
}
