package main

import (
	"bufio"
	"fmt"
	"log"
)

/* Contains status of the communication, from both client and server.
* XXX: We should start with curAck set to -1, which means we have not
* received ACK for 0 block (for client) yet.
* curBlock -> Currently sent block
* curAck -> Currently acknowledged block from the receiver
 */
type myCursor struct {
	fileName string
	curBlock uint16
	curAck   uint16
	op       uint8
	ackZero  bool
	mFile    *myFile
	mReader  *bufio.Reader
}

const (
	BAD_ACK = uint8(1) // Bad ACK packet
	OLD_ACK = uint8(2) // ACK for old data packet
	CUR_ACK = uint8(3) // ACK for the previous data packet
	FIN_ACK = uint8(4) // ACK for the absolute last data packet
)

/* Initialization function for myCursor. */
func (c *myCursor) init(op uint8, fileName string, mfile *myFile,
	isReader bool) {

	c.op = op
	c.fileName = fileName
	c.mFile = mfile
	c.curBlock = 0
	c.curAck = 0
	c.ackZero = false // Special case: negative Ack #

	if isReader {
		c.mReader = c.mFile.startReader()
	}
}

func (c *myCursor) processAck(bufAck []byte, bufLen int) (ackUsed uint8) {

	var n int

	pkt := parsePkt(bufAck, n)
	ackUsed = OLD_ACK

	switch p := pkt.(type) {
	case *myAck:
		log.Println("Got Ack packet\n", p)
		log.Println("Block num in ack :",
			p.blockNum, ", cur block num:",
			c.curBlock)

		/* Set for later ACK only.
		* Check if ackZero is done already.
		 */
		if c.ackZero &&
			c.curAck >= p.blockNum {
			//continue
			ackUsed = OLD_ACK
		} else {
			c.ackZero = true
			c.curAck = p.blockNum
			ackUsed = CUR_ACK

			if bufLen < DATA_SIZE {
				log.Println("bufLen=", bufLen, "\n")
				ackUsed = FIN_ACK
			}
		}
	case *myErr:
		log.Println("Got Err packet\n", p)
		ackUsed = BAD_ACK
	default:
		log.Println("unexpected packet type", p)
		ackUsed = BAD_ACK
	}
	return
}

func (c *myCursor) processData(buf []byte, bufLen int) (datUsed uint8,
	lastAck bool) {

	pkt := parsePkt(buf, bufLen)
	log.Printf("Received pkt: len %d", bufLen)

	datUsed = OLD_ACK
	lastAck = false
	switch p := pkt.(type) {
	case *myDat:
		log.Println("Got Data packet\n")
		log.Println("Block num: ", p.blockNum)
		fmt.Printf("Data: %s\n", string(p.data))

		/* Is this the last block? */
		if bufLen < BLOCK_SIZE {
			lastAck = true
		}
		if c.curAck >= p.blockNum {
			log.Println("Got data pkt already ACKed",
				"cur Ack:", c.curAck,
				", block Num:",
				p.blockNum, "\n")
			datUsed = OLD_ACK
		} else {
			datUsed = CUR_ACK
			c.curBlock = p.blockNum
			c.mFile.writeMyFile(p.readPkt())

			c.curAck = c.curBlock
		}
	default:
		log.Println("unexpected packet type", p)
		datUsed = BAD_ACK
	}
	return
}

func (c *myCursor) getNextBlock() (fileBlock []byte, bufLen int) {

	if c.mReader == nil {
		log.Println("Unexpected Nil reader")
		return
	}
	c.curBlock++
	fileBlock, bufLen = c.mFile.readMyFileBlock(DATA_SIZE, c.mReader)
	return
}
