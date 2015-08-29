package main

import (
	"fmt"
	//	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type myTransmitter struct {
	remoteAddr *net.UDPAddr
	conn       *net.UDPConn
	name       string
	log        *log.Logger
}

func (tx *myTransmitter) txnSend(isServer bool) {

	for i := 0; i < 10; i++ {
		msg := strconv.Itoa(i)
		buf2 := []byte(msg)
		_, err := tx.conn.WriteTo(buf2, tx.remoteAddr)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 1)
	}
}
