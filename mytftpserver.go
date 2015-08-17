package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func handleRequest(tid int, cliAddr *net.UDPAddr, buf []byte, n int,
	mdir *myDir) {

	pkt := parsePkt(buf, n)
	log.Printf("Received pkt: %v", pkt)
	log.Println(" from ", cliAddr)

	switch p := pkt.(type) {
	case *myReq:
		log.Println("Got right packet\n", p)

		worker := new(myWorker)
		worker.run(tid, p, mdir, cliAddr)
	default:
		panic("unknown packet")
	}

	log.Printf("Connection from %v closed.", cliAddr)
}

func main() {

	var mdir myDir

	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(PORT_NUM))
	checkError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)
	defer ServerConn.Close()

	fmt.Println("Server up and listening on port ", PORT_NUM)
	buf := make([]byte, BLOCK_SIZE)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)

		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Received ", string(buf[0:n]), " from ", addr)

		go handleRequest(rand.Intn(65535), addr, buf, n, &mdir)
		checkError(err)
	}
}
