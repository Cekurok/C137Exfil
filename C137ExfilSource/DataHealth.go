package main

import (
	"sort"
	"strconv"
	"time"
)

// Custom Data Health structure
type DataHealth struct {
	OriginalDataSent map[int][]byte
	DataRecv         map[int][]byte
	SendPort         int
	RecvPort         int
	Device           string
	RecvIP           string
	SendIP           string
	ExfilGw          ExfilGateway
}

var (
	lastRun = []int{}
)

// Initalize the Maps for the DH object
func (dh *DataHealth) Init() {
	dh.OriginalDataSent = make(map[int][]byte)
	dh.DataRecv = make(map[int][]byte)
}

// Appending byte slice & Seq # to the custom map
func (dh *DataHealth) AddSent(seq int, data []byte) {
	if _, ok := dh.OriginalDataSent[seq]; ok {
	} else {
		dh.OriginalDataSent[seq] = make([]byte, len(data))

		copy(dh.OriginalDataSent[seq], data)
	}
}

// Appending byte slice & Seq # to the custom map
func (dh *DataHealth) AddRecv(seq int, data []byte) {
	if _, ok := dh.DataRecv[seq]; ok {
		return
	} else {
		dh.DataRecv[seq] = make([]byte, len(data))

		copy(dh.DataRecv[seq], data)
	}

}

// Update the Exfil Gateway Object Ref
func (dh *DataHealth) UpdateExfilGw(ex ExfilGateway) {
	dh.ExfilGw = ex
}

// Update the networking device to use
func (dh *DataHealth) UpdateDevice(device string) {
	dh.Device = device
}

// Update the recv IP
func (dh *DataHealth) UpdateRecvIP(ip string) {
	dh.RecvIP = ip
}

// Update the sending too IP
func (dh *DataHealth) UpdateSendIP(ip string) {
	dh.SendIP = ip
}

// Update the Send too Port
func (dh *DataHealth) UpdateSendPort(port int) {
	dh.SendPort = port
}

// Update the recv Port
func (dh *DataHealth) UpdateRecvPort(port int) {
	dh.RecvPort = port
}

// Request all known missing packets
func (dh *DataHealth) RequestMissingPackets() {

	logData("Sending Requests for Missing Packets..", true, false)

	// Sleep for a set amount of time
	time.Sleep(SLEEPTIME)

	// Loop through each missing packet and send the request packet off.
	for _, seq := range dh.GetMissingSeq() {
		packetGateway.SendPacket(dh.RecvPort, dh.RecvIP, []byte(secretREQ), dh.Device, false, seq, false, false)
	}

}

// Send a request for all data after a specific seq number
func (dh *DataHealth) RequestMissingPacketsFromX() {

	// Sort the list of recv keys
	var keys []int
	for k := range dh.DataRecv {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	// Generate the unique payload
	newPayload := []byte(secretREQ)
	newPayload = append(newPayload, []byte(secretREQALL)...)
	newPayload = append(newPayload, []byte(strconv.Itoa(keys[len(keys)-1]))...)

	// Send off the packet to get the rest of the missing packets
	packetGateway.SendPacket(dh.RecvPort, dh.RecvIP, newPayload, dh.Device, false, "9999", false, false)

}

// Launch the exfil of the data
func (dh *DataHealth) Exfil() {

	// Sort the list of recv keys
	var keys []int
	for k := range dh.DataRecv {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	// For each byte slice of data run the exfil gateway
	for _, i := range keys {
		dh.ExfilGw.RunExfil(dh.DataRecv[i], dh.RecvPort)
	}

}

// Checj to see if we have sent the EOF packet
func (dh *DataHealth) EOFFoundSent() bool {

	for _, data := range dh.OriginalDataSent {
		if checkEOF(data) {
			return true
		}
	}

	return false

}

// Have we currently recieved an EOF packet
func (dh *DataHealth) EOFFound() bool {

	for _, data := range dh.DataRecv {
		if checkEOF(data) {
			return true
		}
	}

	return false

}

// Check for missing data based on seq number
func (dh *DataHealth) IsMissingData() bool {
	var keys []int
	for k := range dh.DataRecv {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	// Start off at our inital offset
	tmp, _ := strconv.Atoi(OFFSETSTART)

	// Loop through each seq # starting at our offset until we hit the last know seq number
	for i := tmp + 1; i <= keys[len(keys)-1]; i++ {
		tmp++
		// If our gen seq # is not present we are missing data
		if contains(keys, tmp) == false {
			return true
		}
	}
	return false
}

// Generate a string slice of all missing seq numbers
func (dh *DataHealth) GetMissingSeq() []string {

	seq := []string{}

	var keys []int
	for k := range dh.DataRecv {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	tmp, _ := strconv.Atoi(OFFSETSTART)

	// Loop through each seq # starting at our offset until we hit the last know seq number
	for i := tmp + 1; i <= keys[len(keys)-1]; i++ {
		tmp++

		if contains(keys, tmp) == false {
			seq = append(seq, strconv.Itoa(tmp))
		}
	}
	return seq
}

// Handle all of the ReTransmissions
func (dh *DataHealth) SendReTransmission(reqAll bool, packet *pPacket) {

	if reqAll {
		// Reset all data past the REQALL Sequence Number
		// Grab the requested Seq Number
		seqint, _ := strconv.Atoi(string(ParserRun(packet.rawData, mapProtocolByPort(packet.toPort))[len([]byte(secretREQ))+len([]byte(secretREQALL)):]))

		// Grab a list of all seq numbers that need to be resent.

		var keys []int
		for k := range dh.OriginalDataSent {
			keys = append(keys, k)
		}

		sort.Ints(keys)

		for i := seqint; i <= keys[len(keys)-1]; i++ {
			seq := strconv.Itoa(i)
			// Check to see if we hit the last seq number if so send the EOF
			if i == keys[len(keys)-1] {
				packetGateway.SendPacket(dh.SendPort, dh.SendIP, dh.OriginalDataSent[i], dh.Device, true, seq, false, false)
			} else {
				packetGateway.SendPacket(dh.SendPort, dh.SendIP, dh.OriginalDataSent[i], dh.Device, false, seq, false, false)
			}

		}
	} else {
		// Pull Seq Number
		seq := string(GetSeqNum(packet.rawData))
		seqIndex, _ := strconv.Atoi(seq)

		// Resend Data based on IP/Port and Map
		if checkEOF(packet.rawData) {
			packetGateway.SendPacket(dh.SendPort, dh.SendIP, dh.OriginalDataSent[seqIndex], dh.Device, true, seq, false, false)
		} else {
			packetGateway.SendPacket(dh.SendPort, dh.SendIP, dh.OriginalDataSent[seqIndex], dh.Device, false, seq, false, false)
		}
	}

}

// Inital Entry Point for ReTransmission
func (dh *DataHealth) ReTransmission() {

	// Wait for the Channel to send back data
	for d := range MainChannel {

		logData("Data Recieved. Parsing & Checking for Call Back...", true, false)

		// Loop through every element in the PacketQ
		for i := 0; i <= d.Size(); i++ {
			packet := d.Pop()

			// Check to see if we have TCP or not
			if packet.PacketType == "TCP" {

				// Identify Secret In Packet
				if IsSecretFound(packet) {

					// Make sure the packet is not from itself
					if packet.sourceIP != GetLocalIP() {

						// Check to see if we are a retrans packet
						if checkSecretREQ(packet.rawData, packet.toPort) {

							// Check to see if we have a Seq ALL REQ
							if checkSecretREQALL(packet.rawData, packet.toPort) {

								dh.SendReTransmission(true, packet)

							} else {

								dh.SendReTransmission(false, packet)
							}

						}

					}

				}

			}
		}
	}

}
