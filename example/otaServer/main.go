package main

import (
	"fmt"
	"github.com/facchinm/ota"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s filepath", os.Args[0])
		os.Exit(1)
	}

	otafile_path := os.Args[1]
	crc := ota.ServeFiles(otafile_path, "")
	ota.StartOTA(crc)
	ota.CloseServer()
}
