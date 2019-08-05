package main

type ExfilGateway struct {
	RecvPort  int
	data      []byte
	ExfilType string
}

// These are persistent variables based on Exfil needs
var (
	ExfilFileReader = FileReader{}
)

// Inital entry point for the Exfil Gateway
func (exfil *ExfilGateway) Init(ExfilType string) {

	// Set the Exfil Type
	exfil.ExfilType = ExfilType

	switch exfil.ExfilType {
	case "FILE":
		ExfilFileReader.Create(EXFIL_FILE, FileBatchSize)
	}
}

// Ran once per packet to handle exfil situations
func (exfil *ExfilGateway) RunExfil(data []byte, port int) {

	// Reset Data
	exfil.data = data
	exfil.RecvPort = port

	// Handle based on exfil type
	switch exfil.ExfilType {
	case "FILE":
		exfil.EXFIL_FILE()
	}

}

// EXFIL MODULE FILE --------
func (exfil *ExfilGateway) EXFIL_FILE() {

	if checkEOF(exfil.data) {
		ExfilFileReader.AppendData(ParserRun(exfil.data, mapProtocolByPort(exfil.RecvPort)), GetSeqNum(exfil.data))
		ExfilFileReader.WriteFile()
	} else {
		ExfilFileReader.AppendData(ParserRun(exfil.data, mapProtocolByPort(exfil.RecvPort)), GetSeqNum(exfil.data))
	}
}
