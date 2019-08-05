package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Packet structure for a single record
type record struct {
	sourceIP      string
	destinationIP string
	frameType     string
	fromPort      int
	toPort        int
	rawData       []byte
	TCPHeader     *layers.TCP
	PacketType    string
}

var (
	handle    *pcap.Handle
	recordSet []record
	currBatch int
)

// Print the individual packets
func (rec *record) printPacket() {
	// Handle Raw Printing of the record
	fmt.Println("Packet Type: ", rec.frameType)
	fmt.Println("Source IP/Port: ", rec.sourceIP, ":", rec.fromPort)
	fmt.Println("Destination IP/Port: ", rec.destinationIP, ":", rec.toPort)
	fmt.Println()
}

// Return thre recordSet
func getRecordSet() []record {
	return recordSet
}

// Add a PacketQ to the channel
func addToChan(channel chan PacketQueue, pq PacketQueue) {
	channel <- pq
}

// Starting Point for Packet Capture
func runCapture(intf string, amountToCapture int, channel chan PacketQueue) {

	// Setup device and currenet batch size
	device := intf
	currBatch = 0

	// Open device
	handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		logData(err.Error(), true, true)
		os.Exit(1)
	}

	// defer the close until after the func returns
	defer handle.Close()

	i := 0

	// Create a new packet source and loop through each packet captured
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {

		// Parse out packet info
		addedPacket := parsePacketInfo(packet)

		// Update our batchsize if we added a packet
		if addedPacket {
			currBatch++
		}
		i++

		// Check to see if we have hit our limit to capture AND we do not have 0 as an amount
		if i == amountToCapture && amountToCapture != 0 {
			break
		}

		// Handle if we need to run forever.
		if currBatch == BATCHSIZE && amountToCapture == 0 {
			// Check to see if our channel was set as it is needed for constant runs
			if channel == nil {
				logData("Channel was not set!", true, true)
				os.Exit(1)
			}
			// Rebuild a PacketQ and append to the Q
			packetQueue := PacketQueue{}
			for _, pack := range recordSet {
				packetQueue.Append(pPacket(pack))
			}

			// Spawn new Go Routine and a send the PacketQ to the recieving proc.
			go addToChan(channel, packetQueue)

			// Reset batches
			recordSet = []record{}
			currBatch = 0

		}
	}
}

// Parse our the packet info
func parsePacketInfo(packet gopacket.Packet) bool {

	// New Record Entry
	rec := record{}

	// Error Handle Int
	errorHandle := 0

	// Handle Ethernet
	errorHandle = parseEthernet(packet, &rec)
	if errorHandle == 1 {
		return false
	}

	// Handle IPV4
	errorHandle = parseIPV4(packet, &rec)
	if errorHandle == 1 {
		return false
	}

	// Handle TCP
	parseTCP(packet, &rec)

	// Handle UDPP
	parseUDP(packet, &rec)

	parseApplication(packet, &rec)
	if errorHandle == 1 {
		return false
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("[Error]: Error decoding some part of the packet:", err)
		return false
	}

	// Add new Rec to recordSet and avoid duplicates
	return addRecord(&rec)
}

// Add new record to the RecordSet
func addRecord(rec *record) bool {

	// Check for duplicates
	if isDuplicateRecord(*rec) {
		return false
	}

	recordSet = append(recordSet, *rec)
	return true
}

// Checks for a duplicate
func isDuplicateRecord(rec record) bool {

	for _, r := range recordSet {
		// check for Source & Dest first as they are the most uncommon to match
		if r.sourceIP == rec.sourceIP && r.destinationIP == rec.destinationIP {
			// We may have a duplicate Check source/Dest IP
			if r.fromPort == rec.fromPort && r.toPort == rec.toPort {
				// We may have a duplicate Check source/Dest IP
				// Check to see if we have the same payload in the packet
				if reflect.DeepEqual(r.rawData, rec.rawData) {
					return true
				}
			}
		}
	}

	return false

}

// Handle the Parsing of the IPV4 Layer
func parseIPV4(packet gopacket.Packet, rec *record) int {

	// Parse IP Header and gather Source/Destination IP as well as Protocol
	// Convert Packet into Layer Type IPV4
	ipLayer := packet.Layer(layers.LayerTypeIPv4)

	// If its not an IPV4 Packet we will just drop it
	if ipLayer != nil {

		// Convert to IPV4 packet Type for parsing
		ip, _ := ipLayer.(*layers.IPv4)

		// Update our Record
		rec.sourceIP = ip.SrcIP.String()
		rec.destinationIP = ip.DstIP.String()

	} else {
		return 1
	}

	return 0
}

// Handle the Parsing of the Application Layer
func parseApplication(packet gopacket.Packet, rec *record) int {

	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		rec.rawData = applicationLayer.Payload()
	} else {
		return 1
	}

	return 0
}

// Handle the Parsing of the UDP Layer
func parseUDP(packet gopacket.Packet, rec *record) int {

	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {

		udp, _ := udpLayer.(*layers.UDP)

		// Update our Record
		rec.fromPort = int(udp.SrcPort)
		rec.toPort = int(udp.DstPort)
		rec.PacketType = "UDP"

	} else {
		return 1
	}

	return 0
}

// Handle the Parsing of the TCP Layer
func parseTCP(packet gopacket.Packet, rec *record) int {

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {

		tcp, _ := tcpLayer.(*layers.TCP)

		// Update our Record
		rec.fromPort = int(tcp.SrcPort)
		rec.toPort = int(tcp.DstPort)
		rec.TCPHeader = tcp
		rec.PacketType = "TCP"

	} else {
		return 1
	}

	return 0
}

// Handle the Parsing of the Ethernet Layer
func parseEthernet(packet gopacket.Packet, rec *record) int {

	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {

		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)

		rec.frameType = ethernetPacket.EthernetType.String()

	} else {
		return 1
	}

	return 0
}
