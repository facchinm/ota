package ota

import (
	"fmt"
	"github.com/gorilla/handlers"
	"hash/crc32"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func getIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && ipnet.IP.String() != "127.0.0.1" && !strings.HasPrefix(ipnet.IP.String(), "169") {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func StartOTA(filename string) bool {

	otafile_path := filename

	//udpAddr, err := net.ResolveUDPAddr("udp4", "192.168.255.255:65500")
	//checkError(err)

	port := 65500

	BROADCAST_IPv4 := net.IPv4(255, 255, 255, 255)

	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: port,
	})
	checkError(err)

	content, _ := ioutil.ReadFile(otafile_path)
	dir, err := ioutil.TempDir("", "ota")
	if err != nil {
		log.Fatal(err)
	}

	padding := make([]byte, ((len(content)/1024)+1)*1024-len(content))
	for i, _ := range padding {
		padding[i] = 0xFF
	}

	content = append(content, padding...)

	defer os.RemoveAll(dir) // clean up

	for i := 0; i < len(content); i = i + 1024 {
		tmpfn := filepath.Join(dir, "otachunk"+strconv.Itoa(i))
		j := 0
		if i+1024 > len(content) {
			j = len(content)
		} else {
			j = i + 1024
		}
		if err := ioutil.WriteFile(tmpfn, content[i:j], 0666); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println(filepath.Join(dir, ""))
	fs := http.FileServer(http.Dir(filepath.Join(dir, "")))

	http.Handle("/", fs)
	go func() {
		http.ListenAndServe(":65201", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	}()

	crcStr := strconv.FormatUint(uint64(crc32.ChecksumIEEE(content)), 10)

	_, err = conn.Write([]byte("OTAUPLOADhttp://" + getIp() + ":65201/" + crcStr))
	checkError(err)

	/* Lets prepare a address at any address at port 10001*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":65500")
	checkError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)
	n, _, err := ServerConn.ReadFromUDP(buf)
	checkError(err)

	fmt.Println(string(buf[0:n]))

	if strings.Contains(string(buf[0:n]), "OK") {
		return true
	} else {
		return false
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}