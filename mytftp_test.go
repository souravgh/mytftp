package main

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func randGen(p []byte) (n int, err error) {

	rand.Seed(time.Now().UnixNano())

	//src := rand.Intn(6) + 1
	for i := range p {
		p[i] = byte(rand.Intn(6) & 0xff)
	}
	return len(p), nil
}

func TestRRQReq(t *testing.T) {

	req := new(myReq)
	req.fileName = "testFile"
	req.op = OP_RRQ

	dataSend := req.sendPkt()

	pkt := parsePkt(dataSend, BLOCK_SIZE)

	switch p := pkt.(type) {
	case *myReq:
		if p.op != OP_RRQ {
			t.Error("Expected", OP_RRQ, " opcode, got ", p.op)
		}
		fileName := strings.Trim(p.fileName, "\x00")
		if req.fileName != fileName {
			t.Error("Expected fileName:", req.fileName, "got :",
				p.fileName)
		}
	default:
		t.Error("Expected req packet, got ", p)

	}
}

func TestWRQReq(t *testing.T) {

	req := new(myReq)
	req.fileName = "test"
	req.op = OP_WRQ

	dataSend := req.sendPkt()

	pkt := parsePkt(dataSend, BLOCK_SIZE)

	switch p := pkt.(type) {
	case *myReq:
		if p.op != OP_WRQ {
			t.Error("Expected", OP_WRQ, " opcode, got ", p.op)
		}
		fileName := strings.Trim(p.fileName, "\x00")

		if req.fileName != fileName {
			t.Error("Expected fileName:", req.fileName, "got :",
				fileName)
		}
	default:
		t.Error("Expected req packet, got ", p)

	}
}

func TestAck(t *testing.T) {

	ack := new(myAck)
	ack.blockNum = 100

	dataSend := ack.sendPkt()

	pkt := parsePkt(dataSend, BLOCK_SIZE)

	switch p := pkt.(type) {
	case *myAck:
		if p.blockNum != 100 {
			t.Error("Expected blockNum:", ack.blockNum, "got: ",
				p.blockNum)
		}
	default:
		t.Error("Expected req packet, got ", p)

	}
}

/* Data packet test with 512 bytes of data. */
func TestDat(t *testing.T) {

	buf := make([]byte, DATA_SIZE)

	n, err := randGen(buf)
	if err != nil {
		t.Error("Error in random generation", err)
	}
	if n != DATA_SIZE {
		t.Error("n expected ", DATA_SIZE, " got :", n)
	}

	datab := new(myDat)
	datab.crtPkt(100, buf, DATA_SIZE)

	dataSend := datab.sendPkt()

	pkt := parsePkt(dataSend, BLOCK_SIZE)

	switch p := pkt.(type) {
	case *myDat:
		if p.blockNum != 100 {
			t.Error("Expected blockNum:", datab.blockNum, "got: ",
				p.blockNum)
		}
		if bytes.Compare(p.data, datab.data) != 0 {
			t.Error("Data bytes do not match")

		}

	default:
		t.Error("Expected req packet, got ", p)

	}
}

/* Data packet test with less than 512 bytes of data. */
func TestDat2(t *testing.T) {

	dataSize := DATA_SIZE - 3
	buf := make([]byte, dataSize)

	n, err := randGen(buf)
	if err != nil {
		t.Error("Error in random generation", err)
	}
	if n != dataSize {
		t.Error("n expected ", dataSize, " got :", n)
	}

	datab := new(myDat)
	datab.crtPkt(100, buf, dataSize)

	dataSend := datab.sendPkt()

	pkt := parsePkt(dataSend, dataSize+4)

	switch p := pkt.(type) {
	case *myDat:
		if p.blockNum != 100 {
			t.Error("Expected blockNum:", datab.blockNum, "got: ",
				p.blockNum)
		}
		if bytes.Compare(p.data, datab.data) != 0 {
			t.Error("Data bytes do not match")

		}

	default:
		t.Error("Expected req packet, got ", p)

	}
}

/* In-memory file reader/writer verification. */
func TestMyFile1(t *testing.T) {

	fileSize := DATA_SIZE
	buf := make([]byte, fileSize)

	n, err := randGen(buf)
	if err != nil {
		t.Error("Error in random generation", err)
	}
	if n != fileSize {
		t.Error("n expected ", fileSize, " got :", n)
	}

	buf1 := buf

	mFile := new(myFile)
	mFile.fileName = "testFile1.dat"
	mFile.writeMyFile(buf)

	buf2 := readTestfile(mFile, fileSize)

	buf3 := readTestfile(mFile, fileSize)

	if bytes.Compare(buf2, buf1) != 0 {
		t.Error("Data bytes of readfile and original buf do not match")
	}

	if bytes.Compare(buf2, buf3) != 0 {
		t.Error("Data bytes from two reads of same file do not match")
	}
}

func readTestfile(mfile *myFile, fileSize int) []byte {

	mReader1 := mfile.startReader()

	buf, _ := mfile.readMyFileBlock(fileSize, mReader1)

	return buf
}
