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
	temp_dir, err := ota.PrepareChunks(otafile_path)
	checkError(err)

	// start in another thread!!!!
	go ota.StartHTTPServer(temp_dir)

	err = ota.SendOTAUDPBroadcast(ota.Crc32Str)
	checkError(err)

	res, err := ota.ReadUDPResponse()
	checkError(err)

	ota.RemoveTempFiles(temp_dir)

	if res == true {
		fmt.Println("Update OK!")
	} else {
		fmt.Println("Update KO :(")
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}
