package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {

	op := "rw"
	if len(os.Args) > 1 {
		op = os.Args[1]
	}

	var fileName string
	if len(os.Args) > 2 {
		fileName = os.Args[2]
	} else {
		fileName = "testFile1"
	}

	var mFile myFile
	mFile.fileName = fileName
	mFile.truncMyFile()
	req := new(myReq)
	req.fileName = mFile.fileName
	switch op {
	case "w":
		/* WRQ!! */
		/* Create a dummy File. */

		for i := 0; i < 9; i++ {
			msg := ""
			for j := 0; j < BLOCK_SIZE-1; j++ {
				msg += strconv.Itoa(i)
			}
			msg += "X"
			buf2 := []byte(msg)
			mFile.writeMyFile(buf2)
		}
		req.op = OP_WRQ

		/* Start communication system. */
		client := new(myClient)

		/* Start processing new file send. */
		client.process(req, &mFile, "127.0.0.1")
	case "r":
		/* RRQ!! */
		req.op = OP_RRQ
		/* Start communication system. */
		client := new(myClient)

		/* Start processing new file send. */
		client.process(req, &mFile, "127.0.0.1")
	default:
		/* WRQ followed by RRQ!! */
		/* Create a dummy File. */

		for i := 0; i < 9; i++ {
			msg := ""
			for j := 0; j < BLOCK_SIZE-1; j++ {
				msg += strconv.Itoa(i)
			}
			msg += "X"
			buf2 := []byte(msg)
			mFile.writeMyFile(buf2)
		}
		req.op = OP_WRQ

		/* Start communication system. */
		client := new(myClient)

		/* Start processing new file send. */
		client.process(req, &mFile, "127.0.0.1")

		buf3 := mFile.readAllFile()

		var mFile2 myFile
		mFile2.fileName = fileName

		req.op = OP_RRQ

		/* Start communication system. */
		client2 := new(myClient)

		/* Start processing new file send. */
		client2.process(req, &mFile2, "127.0.0.1")

		buf4 := mFile2.readAllFile()

		if bytes.Compare(buf3, buf4) != 0 {
			fmt.Println("Data read/write mismatch!\n")
		} else {
			fmt.Println("Data read/write match!\n")
		}
	}

}
