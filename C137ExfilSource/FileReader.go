package main

import (
	"os"
	"sort"
)

// File Reader Custom Structure
type FileReader struct {
	FileHandle *os.File
	fileName   string
	batchSize  int
	data       []byte
	sortedData map[int][]byte
}

// Create the inital reader for future use.
func (fr *FileReader) Create(filename string, batchSize int) {

	// Set filename
	fr.fileName = filename

	// Init Byte Slice
	fr.data = []byte{}

	// Init Map
	fr.sortedData = make(map[int][]byte)

	// set BatchSize for Writting
	fr.batchSize = batchSize

	// Open the File for writing
	fr.FileHandle, err = os.Create(fr.fileName)

	// Check to see if there is an error
	if err != nil {
		logData(err.Error(), true, true)
		os.Exit(1)
	}

}

// Appends a copy of the byte slice to the custom map
func (fr *FileReader) AppendData(data []byte, seqNum []byte) {
	fr.sortedData[seqToInt(seqNum)] = make([]byte, len(data))
	copy(fr.sortedData[seqToInt(seqNum)], data)
}

// Sorts the map into a correctly ordered byte slice for writting.
func (fr *FileReader) BuildData() {

	// Grab a list of ordered keys from the map
	var keys []int
	for k := range fr.sortedData {
		keys = append(keys, k)
	}

	// Sory the keys
	sort.Ints(keys)

	// Append the data to the finaly byte slice
	for _, k := range keys {
		fr.data = append(fr.data, fr.sortedData[k]...)
	}

}

// Write the file to disk
func (fr *FileReader) WriteFile() {

	// defer the close until after this finished
	defer fr.FileHandle.Close()

	// Build the Data from SortedMap
	fr.BuildData()

	// Handle any error while writting data
	if _, err := fr.FileHandle.Write(fr.data); err != nil {
		logData(err.Error(), true, true)
		os.Exit(1)
	}

	logData("File Was Correctly Written..", true, false)
	fr.FileHandle.Close()
}
