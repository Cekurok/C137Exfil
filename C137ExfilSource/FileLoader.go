package main

import (
	"io"
	"os"
)

type FileLoader struct {
	FileHandler *os.File
	FileBytes   []byte
	batchSize   int
}

// Load and init the file to be read from
func (fl *FileLoader) LoadFile(filename string, batchSize int) {

	// Open the File and set it to the handler
	fl.FileHandler, err = os.Open(filename)

	// Set our batch size
	fl.batchSize = batchSize

	// Initalize our Byte Slice
	fl.FileBytes = make([]byte, fl.batchSize)

	// Check to see if we have an error
	if err != nil {
		logData(err.Error(), true, true)
		os.Exit(1)
	}
}

// Pull the data in specific batch sizes
func (fl *FileLoader) NextBatch() []byte {

	// Grab The Batch
	batch, err := fl.FileHandler.Read(fl.FileBytes)

	// Check to see if we are at EOF
	// Return Empty Slice - ONLY time this happens
	if err == io.EOF {
		fl.FileHandler.Close()
		return []byte{}
	}

	return fl.FileBytes[:batch]
}
