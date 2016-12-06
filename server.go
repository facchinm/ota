package ota

import (
	"fmt"
	"github.com/gorilla/handlers"
	"hash/crc32"
	"io/ioutil"
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

func PrepareChunks(filename string) (string, string, error) {
	content, _ := ioutil.ReadFile(filename)
	dir, err := ioutil.TempDir("", "ota")
	if err != nil {
		return "", "", err
	}

	padding := make([]byte, ((len(content)/1024)+1)*1024-len(content))
	for i, _ := range padding {
		padding[i] = 0xFF
	}

	content = append(content, padding...)

	for i := 0; i < len(content); i = i + 1024 {
		tmpfn := filepath.Join(dir, "otachunk"+strconv.Itoa(i))
		j := 0
		if i+1024 > len(content) {
			j = len(content)
		} else {
			j = i + 1024
		}
		if err := ioutil.WriteFile(tmpfn, content[i:j], 0666); err != nil {
			return "", "", err
		}
	}

	crcStr := strconv.FormatUint(uint64(crc32.ChecksumIEEE(content)), 10)

	return dir, crcStr, nil
}

func SendOTAUDPBroadcast(crc32Str string) error {
	port := 65500

	BROADCAST_IPv4 := net.IPv4(255, 255, 255, 255)

	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: port,
	})

	fmt.Printf("Sending UDP broadcast on port %d\n", port)

	_, err = conn.Write([]byte("OTAUPLOADhttp://" + getIp() + ":65201/" + crc32Str))
	return err
}

func StartHTTPServer(directory string) error {
	fmt.Println("Saving chunks in " + filepath.Join(directory, ""))
	fs := http.FileServer(http.Dir(filepath.Join(directory, "")))

	http.Handle("/", fs)
	http.ListenAndServe(":65201", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))

	return nil
}

func ReadUDPResponse() (bool, error) {
	/* Lets prepare a address at any address at port 10001*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":65500")
	if err != nil {
		return false, err
	}

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		return false, err
	}
	defer ServerConn.Close()

	buf := make([]byte, 1024)
	n, _, err := ServerConn.ReadFromUDP(buf)
	if err != nil {
		return false, err
	}

	fmt.Println(string(buf[0:n]))

	if strings.Contains(string(buf[0:n]), "OK") {
		return true, nil
	} else {
		return false, nil
	}

}

func RemoveTempFiles(dir string) {

	os.RemoveAll(dir) // clean up temp files
}
