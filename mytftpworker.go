package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

/*
* Main processing unit for tftp. It includes processing of both server and
* client side.
 */

const (
	TIMEOUT = 600
)

const (
	PORT_NUM = 6000
)

type myConnector struct {
	remoteAddr *net.UDPAddr
	conn       *net.UDPConn
	log        *log.Logger
	status     myCursor
}

/*
* File Sender routine.
 */
func (tx *myConnector) transmitter(isServer bool, op uint8,
	fileName string, mFile *myFile) {

	tx.status.init(op, fileName, mFile, true)
	bufAck := make([]byte, BLOCK_SIZE)

	var fileBlock []byte
	bufLen := DATA_SIZE

	for {
		if !isServer || tx.status.curBlock != 0 {

			n, addr, err := tx.timedReadFromUDP(bufAck)

			if err != nil {
				fmt.Println("error receiving Ack", err)
				return
			}

			/* temp error. */
			if n == 0 {
				if tx.status.curAck == 0 {
					/* Must get first ACK for client.*/
					continue
				}
			} else {
				log.Printf("Received ack: ")
				log.Println(" from ", addr)

				/* No error, got an ACK!. */
				if !isServer && tx.status.curAck == 0 {
					tx.remoteAddr = addr
				}

				ackUsed := tx.status.processAck(bufAck, bufLen)
				switch {
				case ackUsed == BAD_ACK:
					return
				case ackUsed == OLD_ACK:
					continue
				case ackUsed == FIN_ACK:
					return
				default:
				}
			}
		}

		/*
		*Increment cursor block# when we are up to
		*date. CUR_ACK does not mean we successfully
		*sent last block. We may need to try again.
		 */
		if tx.status.curBlock == tx.status.curAck {
			tx.status.curBlock++
			fileBlock, bufLen = tx.status.mFile.readMyFileBlock(
				DATA_SIZE, tx.status.mReader)
		}

		dataPkt := new(myDat)
		fmt.Printf("Data len = %d\n", bufLen)
		dataPkt.crtPkt(tx.status.curBlock, fileBlock, bufLen)
		fmt.Printf("Data: %s\n", string(dataPkt.data))

		n, err := tx.timedWriteTo(dataPkt.sendPkt())

		if err != nil {
			log.Println("Error sending Data for !",
				tx.status.curBlock, "remote Addr",
				tx.remoteAddr)
			return
		}
		if n == 0 {
			log.Println("Could not send Data for !",
				tx.status.curBlock, "remote Addr",
				tx.remoteAddr)
		}
	}
}

/*
* File receiver routine
 */
func (rx *myConnector) receiver(isServer bool, op uint8,
	fileName string, mFile *myFile) {

	rx.status.init(op, fileName, mFile, false)
	lastAck := false

	for {
		// Send the first ACK from the server for block 0,
		// followed by remaining acks.
		if isServer || rx.status.curAck != 0 {
			ack := new(myAck)
			ack.blockNum = rx.status.curAck

			n, err := rx.timedWriteTo(ack.sendPkt())
			if err != nil {
				log.Println("Could not send Ack:",
					rx.status.curAck)
				return
			}
			if n == 0 {
				log.Println("Temp error sending Ack:",
					rx.status.curAck)
				/* Must retry sending the first ACK & last Ack
				 */
				if rx.status.curAck == 0 || lastAck {
					continue
				}
			}
		}

		// Quit, the last one was the final ack
		if lastAck {
			break
		}

		buf := make([]byte, BLOCK_SIZE)
		n, addr, err := rx.timedReadFromUDP(buf)

		if err != nil {
			fmt.Println("error receiving Data", err)
			return
		}

		/* temp error. */
		if n == 0 {
			fmt.Println("temp error receiving Data", err)
			continue
		}

		if !isServer && rx.status.curBlock == 0 && err == nil {
			rx.remoteAddr = addr
		}

		log.Printf("Received pkt from ", rx.remoteAddr)

		/* Process Received Data packet. */
		var datUsed uint8
		datUsed, lastAck = rx.status.processData(buf, n)
		if datUsed == BAD_ACK {
			return
		}
	}
}

/* This is for the client. */
type myClient struct {
	myConnector
}

/* Client request sending routine. */
func (client *myClient) process(req *myReq, mFile *myFile, ip string) {

	servAddr := net.UDPAddr{
		Port: PORT_NUM,
		IP:   net.ParseIP(ip),
	}

	myAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	client.conn, err = net.ListenUDP("udp", myAddr)
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.conn.WriteToUDP(req.sendPkt(), &servAddr)
	if err != nil {
		fmt.Println(req.fileName, err)
	}
	fmt.Println("Remote Address :", servAddr)
	fmt.Println("Local Address :", *myAddr)

	switch req.op {
	case OP_WRQ:
		client.transmitter(false, req.op, req.fileName, mFile)
	case OP_RRQ:
		client.receiver(false, req.op, req.fileName, mFile)
	}
}

/* This is for a new thread of the server. */
type myWorker struct {
	myConnector
}

/* Server request handline routine. */
func (worker *myWorker) run(tid int, req *myReq, mDir *myDir,
	cliAddr *net.UDPAddr) {

	var mFile *myFile

	log.Println("Client", cliAddr, "connected.")

	myAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	worker.conn, err = net.ListenUDP("udp", myAddr)
	if err != nil {
		log.Fatal(err)
	}

	worker.remoteAddr = cliAddr

	switch req.op {
	case OP_WRQ:
		mFile = mDir.findFile(req.fileName)

		if mFile != nil {
			errp := new(myErr)
			errp.errCode = ER_EEXIST
			errp.errMsg = "File already exists."
			log.Println("File ", req.fileName, "already exists.")
			_, err = worker.conn.WriteTo(errp.sendPkt(),
				worker.remoteAddr)
			if err != nil {
				log.Println("Could not send first err packet!")
			}
			return
		}

		mFile = new(myFile)
		mFile.fileName = req.fileName
		mDir.addFile(mFile)
		worker.receiver(true, req.op, req.fileName, mFile)
	case OP_RRQ:
		mFile = mDir.findFile(req.fileName)

		if mFile == nil {
			log.Println("File ", req.fileName, "not found")

			errp := new(myErr)
			errp.errCode = ER_ENOENT
			errp.errMsg = "File not found."
			log.Println("File ", req.fileName, "not found.")
			_, err = worker.conn.WriteTo(errp.sendPkt(),
				worker.remoteAddr)
			if err != nil {
				log.Println("Could not send first err packet!")
			}
			return
		}
		worker.transmitter(true, req.op, req.fileName, mFile)
	default:
		log.Fatal("Unknown op", worker.status.op)
	}
}

/* Handle timeouts during UDP read. */
func (c *myConnector) timedReadFromUDP(buf []byte) (nm int,
	raddr *net.UDPAddr, rerr error) {

	c.conn.SetDeadline(time.Now().Add(TIMEOUT * time.Second))

	n, addr, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		if nerr, ok := err.(net.Error); ok &&
			nerr.Temporary() {
			/* Handle temporary errors. */
			fmt.Println("temp error receiving Data", err)
			time.Sleep(time.Second * 1)
			rerr = nil
			nm = 0
		} else {
			fmt.Println("error receiving Data", err)
			fmt.Println(err)
			rerr = err
			nm = 0
		}
	} else {
		nm = n
		raddr = addr
		rerr = err
	}
	return
}

/* Handle timeouts during UDP write. */
func (c *myConnector) timedWriteTo(buf []byte) (nm int, rerr error) {
	c.conn.SetDeadline(time.Now().Add(TIMEOUT * time.Second))

	n, err := c.conn.WriteTo(buf, c.remoteAddr)

	if err != nil {
		log.Println("Could not send Data for !",
			"remote Addr", c.remoteAddr)
		/* Handle temporary errors. */
		if nerr, ok := err.(net.Error); ok &&
			nerr.Temporary() {
			fmt.Println(
				"temp error sending Data",
				err)
			time.Sleep(time.Second * 1)
			rerr = nil
			nm = 0
		} else {
			fmt.Println("error receiving Data", err)
			fmt.Println(err)
			rerr = err
			nm = 0
		}
	} else {
		nm = n
		rerr = err
	}
	return
}
