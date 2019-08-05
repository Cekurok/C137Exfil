package main

// Default entry point for parser engine
func ParserRun(data []byte, protocol string) []byte {
	// Check for each protocol and run the correct parser in that module
	if protocol == "HTTP" {
		return ParseHTTP(data)
	}

	return data
}

// Handle HTTP Parsing
func ParseHTTP(data []byte) []byte {
	return ParseHTTPData(data, true)
}
