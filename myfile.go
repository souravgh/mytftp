package main

import (
	"bufio"
	"bytes"
	"fmt"
)

/*
* My In-memory ad-hoc file system. Each file is represented by myFile.
* The directory is represented as a slice of mFile(s).
 */
type myFile struct {
	fileName string
	len      int
	dat      bytes.Buffer
}

func (mfile *myFile) truncMyFile() {
	mfile.dat.Truncate(0)
	mfile.len = 0
}

func (mfile *myFile) writeMyFile(data []byte) {
	mfile.len += len(data)
	mfile.dat.Write(data)
}

func (mfile *myFile) startReader() (r *bufio.Reader) {
	readBuffer := bytes.NewBuffer(mfile.dat.Bytes())
	r = bufio.NewReader(readBuffer)

	return r
}

func (mfile *myFile) readMyFileBlock(dataSize int,
	r *bufio.Reader) (fileBlock []byte, bufLen int) {

	fileBlock = make([]byte, dataSize)

	bufLen, err := r.Read(fileBlock)
	if err != nil {
		fmt.Println("Error getting read:", err, "\n")
	}
	return
}

func (mfile *myFile) readAllFile() []byte {

	mReader1 := mfile.startReader()

	buf, _ := mfile.readMyFileBlock(mfile.len, mReader1)

	return buf
}

type myDir struct {
	files []*myFile
}

func (mdir *myDir) findFile(fileName string) (mfile *myFile) {
	var file *myFile

	for i := range mdir.files {
		name := mdir.files[i].fileName

		if name == fileName {
			file = mdir.files[i]
			break
		}
	}
	return file
}

func (mdir *myDir) addFile(mfile *myFile) {
	mdir.files = append(mdir.files, mfile)
}
