package main

import (
	"bytes"
	"encoding/binary"
	"strings"
)

/*
* Packet structures to marshall and unmarshall data across the wire.
 */
const (
	OP_RRQ = uint8(1) // Read request
	OP_WRQ = uint8(2)
	OP_DAT = uint8(3)
	OP_ACK = uint8(4)
	OP_ERR = uint8(6)
)

type myPkt interface {
	sendPkt() []byte
	recvPkt([]byte, int)
}

const (
	BLOCK_SIZE = 516
	DATA_SIZE  = 512
)

/*
   Read/Write request packet.
*/
type myReq struct {
	fileName string
	op       uint8
}

func (req *myReq) sendPkt() []byte {
	return sendReqPkt(req.fileName, req.op)
}

func (req *myReq) recvPkt(data []byte, n int) {
	req.fileName, req.op = recvReqPkt(data, n)
}

func sendReqPkt(fileName string, op uint8) []byte {
	buffer := &bytes.Buffer{}

	buffer.WriteByte(0x0)
	buffer.WriteByte(op)
	buffer.WriteString(fileName)
	buffer.WriteByte(0x0)
	buffer.WriteString("octet")
	buffer.WriteByte(0x0)

	return buffer.Bytes()
}

func recvReqPkt(data []byte, n int) (fileName string, op uint8) {
	buffer := bytes.NewBuffer(data)

	_, err := buffer.ReadByte()
	if err != nil {
		panic(err)
	}

	op, err = buffer.ReadByte()
	if err != nil {
		panic(err)
	}

	fileName, err = buffer.ReadString(0x0)
	if err != nil {
		panic(err)
	}
	fileName = strings.TrimSpace(fileName)

	//Ignore mode, since it will always be in octet
	return
}

/*
   ACK packet
*/
type myAck struct {
	blockNum uint16
}

func (ack *myAck) sendPkt() []byte {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(0x0)
	buffer.WriteByte(OP_ACK)
	binary.Write(buffer, binary.BigEndian, ack.blockNum)

	return buffer.Bytes()
}

func (ack *myAck) recvPkt(data []byte, n int) {
	buffer := bytes.NewBuffer(data)

	_, err := buffer.ReadByte()
	if err != nil {
		panic(err)
	}

	_, err = buffer.ReadByte()
	if err != nil {
		panic(err)
	}

	err = binary.Read(buffer, binary.BigEndian, &ack.blockNum)
	if err != nil {
		panic(err)
	}
}

/*
   DATA packet.
*/
type myDat struct {
	blockNum uint16
	size     int
	data     []byte
}

func (dat *myDat) getSize() int {
	return len(dat.data)
}

func (dat *myDat) crtPkt(bNum uint16, buf []byte, size int) {
	dat.blockNum = bNum
	dat.data = buf
	dat.size = size
}

func (dat *myDat) sendPkt() []byte {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(0x0)
	buffer.WriteByte(OP_DAT)
	binary.Write(buffer, binary.BigEndian, dat.blockNum)
	buffer.Write(dat.data)
	buffer.Truncate(dat.size + 4)

	return buffer.Bytes()
}

func (dat *myDat) readPkt() []byte {
	buffer := &bytes.Buffer{}
	buffer.Write(dat.data)
	buffer.Truncate(dat.size)

	return buffer.Bytes()
}

func (dat *myDat) recvPkt(data []byte, n int) {
	buffer := bytes.NewBuffer(data)
	buffer.Truncate(n)

	_, err := buffer.ReadByte()
	if err != nil {
		panic(err)
	}

	_, err = buffer.ReadByte()
	if err != nil {
		panic(err)
	}

	err = binary.Read(buffer, binary.BigEndian, &dat.blockNum)
	if err != nil {
		panic(err)
	}

	Len := buffer.Len()
	dat.size = Len
	dat.data = data[4:]
}

const (
	ER_ENOENT  = uint16(1) // File not found.
	ER_EACCESS = uint16(2) // Access violation.
	ER_ENOSPC  = uint16(3) // Disk full or allocation exceeded.
	ER_EBADOP  = uint16(4) // Illegal TFTP operation.
	ER_ETID    = uint16(5) // Unknown transfer ID.
	ER_EEXIST  = uint16(6) // File already exists.
	ER_ENOUSER = uint16(7) // No such user.
)

/*
   ERROR packet.
*/
type myErr struct {
	errCode uint16
	errMsg  string
}

func (err *myErr) sendPkt() []byte {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(0x0)
	buffer.WriteByte(OP_ERR)
	binary.Write(buffer, binary.BigEndian, err.errCode)
	buffer.WriteString(err.errMsg)
	buffer.WriteByte(0x0)

	return buffer.Bytes()
}

func (err *myErr) recvPkt(data []byte, n int) {
	buffer := bytes.NewBuffer(data)

	_, e := buffer.ReadByte()
	if e != nil {
		panic(e)
	}

	_, e = buffer.ReadByte()
	if e != nil {
		panic(e)
	}
	e = binary.Read(buffer, binary.BigEndian, &err.errCode)
	if e != nil {
		panic(e)
	}

	errMsg, e := buffer.ReadString(0x0)
	if e != nil {
		panic(e)
	}
	err.errMsg = strings.TrimSpace(errMsg)

}

func parsePkt(data []byte, n int) myPkt {
	var pkt myPkt

	switch data[1] {
	case OP_RRQ:
		pkt = new(myReq)
		pkt.recvPkt(data, n)
	case OP_WRQ:
		pkt = new(myReq)
		pkt.recvPkt(data, n)
	case OP_DAT:
		pkt = new(myDat)
		pkt.recvPkt(data, n)
	case OP_ACK:
		pkt = new(myAck)
		pkt.recvPkt(data, n)
	case OP_ERR:
		pkt = new(myErr)
		pkt.recvPkt(data, n)
	default:
		panic("Unknown packet")
	}
	return pkt
}
