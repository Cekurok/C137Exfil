package main

func main() {

	// Print the Banner for the Application
	printBanner()

	// Parse out the CommandLine Arguments
	cmd := CommandArgs{}
	cmd.Parse()

	// Create Exfil Handler
	exfilHandler := ExfilHandler{}

	// Run Exfil Handler
	exfilHandler.Run()
}
