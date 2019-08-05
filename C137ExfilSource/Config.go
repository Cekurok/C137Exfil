package main

import "time"

const (
	// CHANGE THESE FOR A NEW SIGNATURE!
	SIGNATURE         = 12444568568
	BATCHSIZE         = 500
	PACKETCAPTURESIZE = 1000
	DEFAULTPROTOCOL   = "HTTP"
	SLEEPTIME         = 2 * time.Second
)

var (
	avilProtocols = []string{"HTTP"} // UPDATE BASED AVAILABLE MODULES
	avilExfil     = []string{"FILE"} // UPDATE BASED AVAILABLE MODULES
	protMap       = map[string]int{
		"HTTP":   80,
		"SMTP":   25,
		"HTTPS":  443,
		"FTP":    20,
		"FTPD":   21,
		"TELNET": 23,
		"IMAP":   143,
		"SSH":    22,
		"DNS":    53,
		"DHCP":   67,
		"POP3":   110,
	}

	OFFSETSTART = "1000" // Must be 1000 or more!

	snapshotLen   int32 = 1024
	promiscuous   bool  = false
	err           error
	timeout       time.Duration = 30 * time.Second
	MainChannel                 = make(chan PacketQueue)
	silent                      = false
	FileBatchSize               = 500 // DO NOT CHANGE ABOVE 1000! TCP MTU limit is 1500 per packet

	// EXFIL MODULES ----------------------------------------------

	EXFIL_FILE = "test.png" // Change this for the filename to be dumped to disk

	// PROTOCOL MODULES ----------------------------------------------

	// HTTP START

	HTTPBatchSize = 1000

	GetURL    = "/"
	PostURL   = "/uploadImage"
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"

	// GET
	GetRequest = map[string]string{
		"GET":             GetURL + " HTTP/1.1",
		"Connection":      "keep-alive",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "en-US,en;q=0.9",
		"User-Agent":      UserAgent,
	}

	// POST
	PostRequest = map[string]string{
		"POST":           PostURL + " HTTP/1.1",
		"Connection":     "keep-alive",
		"Accept":         "*/*",
		"Content-Length": "1024",
		"Content-Type":   "application/x-www-form-urlencoded",
		"User-Agent":     UserAgent,
	}

	// HTTP END
)
