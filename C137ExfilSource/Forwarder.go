package main

import "fmt"

var (
	currentPacketSet = []record{}
)

// Handle the Forwarder logic
func StartForwarder(channel chan PacketQueue, nextHop string, device string, remotePort int, pg PacketGateway, cmd CommandArgs, dh DataHealth) {

	logData("Starting Forwarder....", true, false)

	// Wait for each PacketQ to be sent to the channel
	for d := range channel {

		for i := 0; i <= d.Size(); i++ {
			packet := d.Pop()

			// Handle if it is TCP
			if packet.PacketType == "TCP" {

				// Identify Secret In Packet
				if IsSecretFound(packet) {

					// Make sure the packet is not from itself
					if packet.sourceIP != GetLocalIP() {

						logData("Data Recieved. Parsing & Forwarding...", true, false)

						// Handle if the packet is a Retrans
						if checkSecretREQ(packet.rawData, packet.toPort) {
							// Check to see if we have a Seq ALL REQ
							if checkSecretREQALL(packet.rawData, packet.toPort) {
								fmt.Println("Got REQALL ReTrans")
								dh.SendReTransmission(true, packet)
							} else {
								fmt.Println("Got REQ ReTrans")
								dh.SendReTransmission(false, packet)
							}
						} else {

							// Handle if it just needs to be pushed to the next hop
							// Get the Seq Num
							seqNum := string(GetSeqNum(packet.rawData))

							// Handle Update to Recv Info
							dh.UpdateRecvIP(packet.sourceIP)
							dh.UpdateRecvPort(packet.toPort)

							//Add to Recv in DH
							dh.AddRecv(seqToInt([]byte(seqNum)), packet.rawData)

							fmt.Println("Forwarding Packet!")

							// Send off the data to the next hop
							if checkEOF(packet.rawData) {
								packetGateway.SendPacket(mapPortByProtocol(cmd.getRemoteProtocol()), cmd.getNextHop(), packet.rawData, cmd.getInterface(), true, seqNum, true, true)
							} else {
								packetGateway.SendPacket(mapPortByProtocol(cmd.getRemoteProtocol()), cmd.getNextHop(), packet.rawData, cmd.getInterface(), false, seqNum, true, true)
							}

							// Update DH Sent
							dh.AddSent(seqToInt([]byte(seqNum)), packet.rawData)
							if checkEOF(packet.rawData) {
								fmt.Println("REcieved EOF: ", seqNum)
							}

							// Do we have all of the Data from recv?
							if dh.IsMissingData() == false && dh.EOFFound() {
								// We have everything
								// TODO Add a reset to DH?
								continue
							}

							// Handle if we Dont have EOF but have no missing data
							if dh.IsMissingData() == false && dh.EOFFound() == false {
								fmt.Println("Sending MissingX")
								dh.RequestMissingPacketsFromX()
							}

							// get missing data if we dont have everything
							if dh.IsMissingData() {
								fmt.Println("Asking for Missing Packet!")
								// Request the Data again
								dh.RequestMissingPackets()
							}
						}

					}

				}

			}
		}
	}

}
