package main

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
)

// Secret Values
var (
	secret          = []int{}
	secretEgg       = ""
	secretEggEOF    = ""
	generatedSecret = false
	secretREQ       = ""
	secretREQALL    = ""
)

// Check to see if the secrets are found inside of a packet
func IsSecretFound(packet *pPacket) bool {

	// Identify Secret In Packet
	totalMatches := 0

	// Checking each TCP Flag for a match
	for _, sec := range secret {
		if flags[sec] == "FIN" {
			if packet.TCPHeader.FIN {
				totalMatches++
			}
		} else if flags[sec] == "SYN" {
			if packet.TCPHeader.SYN {
				totalMatches++
			}
		} else if flags[sec] == "RST" {
			if packet.TCPHeader.RST {
				totalMatches++
			}
		} else if flags[sec] == "ACK" {
			if packet.TCPHeader.ACK {
				totalMatches++
			}
		} else if flags[sec] == "URG" {
			if packet.TCPHeader.URG {
				totalMatches++
			}
		} else if flags[sec] == "ECE" {
			if packet.TCPHeader.ECE {
				totalMatches++
			}
		} else if flags[sec] == "CWR" {
			if packet.TCPHeader.CWR {
				totalMatches++
			}
		} else if flags[sec] == "NS" {
			if packet.TCPHeader.NS {
				totalMatches++
			}
		}

		// Check to see if we have the correct amount of flags based on the secret.
		if totalMatches == len(secret) {

			// Check to see if the secret egg is present
			if checkEgg(packet.rawData) {
				return true
			}
		}
	}

	return false
}

// Checks for the custom Request All after X packet.
func checkSecretREQALL(data []byte, port int) bool {

	// Strip our inital data
	strippedData := ParserRun(data, mapProtocolByPort(port))

	// Check length to see if its even there.
	if len(strippedData) >= len([]byte(secretREQ))+len([]byte(secretREQALL)) {
		if string(strippedData[len([]byte(secretREQ)):len([]byte(secretREQ))+len([]byte(secretREQALL))]) == secretREQALL {
			return true
		}
	}

	return false

}

// Checks for the custom Request packet.
func checkSecretREQ(data []byte, port int) bool {

	// Strip our inital data
	strippedData := ParserRun(data, mapProtocolByPort(port))

	// Check length to see if its even there.
	if len(strippedData) >= len([]byte(secretREQ)) {

		if string(strippedData[:len([]byte(secretREQ))]) == secretREQ {
			return true
		}
	}

	return false

}

// Check to see if our egg is present in the raw data
func checkEgg(data []byte) bool {

	// Do we have enough bytes for there to be an egg?
	if len(data) < len([]byte(secretEgg)) {
		return false
	}

	// Calculate the offset to the secret egg and compare it
	if string(data[len([]byte(OFFSETSTART)):len([]byte(secretEgg))+len([]byte(OFFSETSTART))]) == secretEgg {
		return true
	}
	return false
}

// Generate a random number between 2 sets
func random(min int, max int) int {
	return rand.Intn(max-min) + min
}

// RUNS ONLY ONCE!
// Generates the inital secrets for the binary
func generateSecret() {

	// Seed random with our Signature
	rand.Seed(SIGNATURE)

	// Get how many TCP flags we will be using
	randFlagCount := random(1, 6)

	// Generate which TCP flags based on the random amount to be used
	for i := 0; i < randFlagCount; i++ {
		newRand := random(0, 7)

		for contains(secret, newRand) {
			newRand = random(0, 7)
		}
		secret = append(secret, newRand)
	}

	// Generate all secrets required
	secretEgg = strconv.Itoa(random(1, SIGNATURE))
	secretEggEOF = strconv.Itoa(random(1, SIGNATURE))
	secretREQ = strconv.Itoa(random(1, SIGNATURE))
	secretREQALL = strconv.Itoa(random(1, SIGNATURE))

	// Set this to true so we dont generate again
	generatedSecret = true
}

// Strip the secret egg from our raw data
func stripSecretEgg(data []byte) []byte {
	return data[len([]byte(OFFSETSTART))+len([]byte(secretEgg)):]
}

// Strip the EOF egg from raw data
func stripSecretEggEOF(data []byte) []byte {
	return data[:len(data)-len([]byte(secretEggEOF))]
}

// Calc the MD5 of a byte slice
func GetMD5(data []byte) string {
	hashObj := md5.New()
	hash := hashObj.Sum(data)[:16]
	return hex.EncodeToString(hash)
}
