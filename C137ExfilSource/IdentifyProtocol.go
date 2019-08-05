package main

import (
	"strings"
)

var (
	protocolMap map[string]int
)

// Entry Point for identification of a protocol
func RunProtocolIdentify(cmd CommandArgs) string {

	logData("Identifying Best Protocol for Exfil to Host: "+cmd.getNextHop()+"...", true, false)

	// Start Packet Capture for packet samples
	runCapture(cmd.getInterface(), PACKETCAPTURESIZE, nil)

	// Sort out the most used Protocols to the Exfil Host based on available protocols
	IdentifyProtocolByIP(cmd.getNextHop(), getRecordSet())

	// Get Most Used Protocol
	return GetMostUsedProtocol()

}

// Identify protocol based on the IP and record set of packets.
func IdentifyProtocolByIP(ip string, recordSet []record) {

	// Init Map
	protocolMap = map[string]int{}

	// Loop through every record AKA packet
	for _, rec := range recordSet {

		//Check to see if the Source IP does not Matche our next Hop
		if rec.sourceIP != ip {
			continue
		}

		// Check to see if the key currently exists
		result := mapProtocolByPort(rec.toPort)
		if _, present := protocolMap[result]; present {
			protocolMap[mapProtocolByPort(rec.toPort)] = protocolMap[mapProtocolByPort(rec.toPort)] + 1
		} else {
			protocolMap[mapProtocolByPort(rec.toPort)] = 1
		}

	}
}

// get the most used protocol.
func GetMostUsedProtocol() string {

	//Use the default protocol to start
	mostUsedProtocol := DEFAULTPROTOCOL
	mostUsedAmt := 0

	// Run through protocol map and validate if anything matches
	for prot, amt := range protocolMap {
		if amt > mostUsedAmt && containsStr(avilProtocols, prot) {
			mostUsedProtocol = prot
			mostUsedAmt = amt
		}
	}
	return mostUsedProtocol

}

// Get the Raw Protocol Map
func GetProtocolMap() map[string]int {
	return protocolMap
}

// Map the protocol by the port number
func mapProtocolByPort(port int) string {

	// Loop through the protocol map and match by port
	for protocol, p := range protMap {
		if p == port {
			return protocol
		}
	}
	return DEFAULTPROTOCOL
}

// Map the port by the protocol number
func mapPortByProtocol(prot string) int {

	// Loop through the protocol map and match by protocol
	for protocol, p := range protMap {
		if strings.ToUpper(prot) == protocol {
			return p
		}
	}
	return 0
}
