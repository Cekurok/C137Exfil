package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Custom Packet Structure to be used in most places
type Packet struct {
	sourceIP         []byte
	destinationIP    []byte
	rawDestinationIP string
	destinationPort  int
	seq              []byte
	flags            map[string]bool
	FIN              bool
	SYN              bool
	RST              bool
	ACK              bool
	URG              bool
	ECE              bool
	CWR              bool
	NS               bool
	device           string
	data             string
	IPLayer          *layers.IPv4
	ETHLayer         *layers.Ethernet
	TCPLayer         *layers.TCP
	rawData          []byte
}

// Set the TCP Flag Options
var (
	flags = []string{"FIN", "SYN", "RST", "ACK", "URG", "ECE", "CWR", "NS"}
)

// Build an inital Empty Packet Object
func BuildBaseTCPPacket(destIP string, rawdevice string, destPort int) Packet {

	p := Packet{}
	p.sourceIP = net.ParseIP(GetLocalIP())
	p.rawDestinationIP = destIP
	p.destinationIP = net.ParseIP(destIP)
	p.device = rawdevice
	p.destinationPort = destPort
	p.seq = []byte{}

	p.CreateBaseFlags()

	return p
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func getLocalMacAddr() string {
	addr := ""
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				// Don't use random as we have a real address
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return addr
}

// Get the remote Mac Address
// TODO - Make it not OS Dependent
func GetRemoteMac(ip string, device string, init bool) string {

	// Only ping the IP once
	if init {
		cmd := exec.Command("ping", "-c", "1", ip)
		_ = cmd.Run()
		return ""
	}

	// Pull from the ARP Table the MAC Address
	cmd := exec.Command("arp", "-n", "-i", device, ip)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err = cmd.Run()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	return strings.Split(string(cmdOutput.Bytes()), " ")[3]
}

// Enable a Specific Flag if needed.
func (packet *Packet) EnableFlag(flag string) {

	switch strings.ToUpper(flag) {
	case "FIN":
		packet.flags["FIN"] = true
	case "SYN":
		packet.flags["SYN"] = true
	case "RST":
		packet.flags["RST"] = true
	case "ACK":
		packet.flags["ACK"] = true
	case "URG":
		packet.flags["URG"] = true
	case "ECE":
		packet.flags["ECE"] = true
	case "CWR":
		packet.flags["CWR"] = true
	case "NS":
		packet.flags["NS"] = true
	default:
		fmt.Println("FLAG ", flag, " Not a Valid TCP Flag!")
	}
}

// Create the base set of flags in the packet as disabled.
func (packet *Packet) CreateBaseFlags() {

	packet.flags = map[string]bool{}

	for _, flag := range flags {
		packet.flags[flag] = false
	}
}

// Update the Seq Number of the packet
func (packet *Packet) UpdateSeqNum(seqNum []byte) {
	packet.seq = seqNum
}

// Compile the packet prior to being sent.
func (packet *Packet) CompilePacket() {

	// Handle Custom Seq Num
	if len(packet.seq) != len([]byte(OFFSETSTART)) {
		logData("Cannot Send packet without Seq Number!", true, true)
		os.Exit(1)
	}

	// Build each layer of the packet by hand.
	packet.BuildETHLayer()
	packet.BuildIPLayer()
	packet.BuildTCPLayer()
}

// Add the raw packet data to be sent
func (packet *Packet) addData(data []byte) {
	packet.rawData = data
}

// Parse out the Mac address based on a string version
func (packet *Packet) parseMAC(mac string) net.HardwareAddr {

	hdwr, _ := net.ParseMAC(mac)
	return hdwr
}

// Build out the IP layer of the packet
func (packet *Packet) BuildIPLayer() {
	ipLayer := &layers.IPv4{
		SrcIP:    packet.sourceIP,
		DstIP:    packet.destinationIP,
		Protocol: layers.IPProtocolTCP,
		Version:  4,
		TTL:      128,
	}

	packet.IPLayer = ipLayer
}

// Build out the Ethernet Layer
func (packet *Packet) BuildETHLayer() {
	ethernetLayer := &layers.Ethernet{
		SrcMAC:       packet.parseMAC(getLocalMacAddr()),
		DstMAC:       packet.parseMAC(GetRemoteMac(packet.rawDestinationIP, packet.device, false)),
		EthernetType: layers.EthernetTypeIPv4,
	}

	packet.ETHLayer = ethernetLayer
}

// Build out the TCP Layer
func (packet *Packet) BuildTCPLayer() {
	tcpLayer := &layers.TCP{
		SrcPort: layers.TCPPort(random(1000, 9999)),
		DstPort: layers.TCPPort(packet.destinationPort),
		Window:  65535,
	}

	// Enable the flags based on the secrets generated
	for i := range secret {
		packet.flags[flags[secret[i]]] = true
	}

	// Modify the packet to have these flags
	for key, value := range packet.flags {
		if key == "FIN" && value {
			tcpLayer.FIN = true
		} else if key == "SYN" && value {
			tcpLayer.SYN = true
		} else if key == "RST" && value {
			tcpLayer.RST = true
		} else if key == "ACK" && value {
			tcpLayer.ACK = true
		} else if key == "URG" && value {
			tcpLayer.URG = true
		} else if key == "ECE" && value {
			tcpLayer.ECE = true
		} else if key == "CWR" && value {
			tcpLayer.CWR = true
		} else if key == "NS" && value {
			tcpLayer.NS = true
		}
	}

	tcpLayer.SetNetworkLayerForChecksum(packet.IPLayer)

	packet.TCPLayer = tcpLayer
}

// Sending the packet itself
func (packet *Packet) send(withEOF bool) {

	// Setup inital variables
	var snapshot_len int32 = 1024
	var promiscuous bool = false
	var err error
	var timeout time.Duration = 30 * time.Second
	var handle *pcap.Handle
	var buffer gopacket.SerializeBuffer
	var options gopacket.SerializeOptions

	// Open device
	handle, err = pcap.OpenLive(packet.device, snapshot_len, promiscuous, timeout)
	if err != nil {
		logData(err.Error(), true, true)
		os.Exit(1)
	}
	defer handle.Close()

	// And create the packet with the layers
	buffer = gopacket.NewSerializeBuffer()
	options = gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	// Setup the base payload and add the EOF to it if needed
	// The bytes are added in reverse order for the packet to be send correctly
	payload := []byte{}
	if withEOF {
		payload = append([]byte(secretEggEOF), payload...)
	}

	// Build the rest of the payload
	payload = append(packet.rawData, payload...)
	payload = append([]byte(secretEgg), payload...)
	payload = append(packet.seq, payload...)

	// Do final packet compile to bytes
	gopacket.SerializeLayers(buffer, options,
		packet.ETHLayer,
		packet.IPLayer,
		packet.TCPLayer,
		gopacket.Payload(payload),
	)
	outgoingPacket := buffer.Bytes()

	// send packet off on the wire
	err = handle.WritePacketData(outgoingPacket)
	if err != nil {
		logData(err.Error(), true, true)
	}

}

// Check to see if we hit the EOF in a packet
func checkEOF(data []byte) bool {
	if len(data) <= len([]byte(secretEggEOF)) {
		return false
	}
	if string(data[len(data)-len([]byte(secretEggEOF)):]) == secretEggEOF {
		return true
	}
	return false
}

// Increment the Seq Number based on a string
func IncSeqNum(seq string) string {
	seqInt, _ := strconv.Atoi(seq)
	seqInt++
	return strconv.Itoa(seqInt)
}

// Convert a string version fo the seq num to a byte slice
func seqToByte(seq string) []byte {
	return []byte(seq)
}

// Convert a byte slice version fo the seq num to a string
func seqToString(seq []byte) string {
	return string(seq)
}

// Convert a byte slice version fo the seq num to an int
func seqToInt(seq []byte) int {
	ret, _ := strconv.Atoi(string(seq))
	return ret
}

// Return the seq num from a packet
func GetSeqNum(data []byte) []byte {
	return data[:len([]byte(OFFSETSTART))]
}

// strip the seq num from the packet and return whats left
func stripSeqNum(data []byte) []byte {
	return data[len([]byte(OFFSETSTART)):]
}
