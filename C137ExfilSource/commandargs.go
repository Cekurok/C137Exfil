package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Needs to be updated with the latest list of possible commands
type CommandArgs struct{}

//ONLY CHANGE THESE -------
var (
	// Handle Interface
	NetInterface = flag.String("i", "", "Interface to Listen On For Smart Traffic")

	// Exfil Dynamically
	DynamicRun = flag.Bool("d", false, "Dynamically Exfil to the next hop based on current traffic. Default will use HTTP/80.")

	// Handle Next Hop
	NextHop = flag.String("h", "127.0.0.1", "Next Hop to Exfil the Data Too. Needs to be a host that has another C137EXFIL running!")

	// Custom Port
	Port = flag.String("p", "", "Next Hop's Protocol to Exfil the Data Too. Use -lp for a list of currently supported Protocols.")

	// List Available Protocols
	ListProts = flag.Bool("lp", false, "List of currently supported Protocols.")

	// Input File to be exfiled
	ExfilFile = flag.String("f", "", "File to be Exfiled.")

	// List Available Protocols
	ListExfil = flag.Bool("le", false, "List of currently supported Exfils.")

	// Output File to be exfiled
	ExfilProtocol = flag.String("e", "", "Type of Exfil To Use. See -le for options on Exfils") // TODO CHANGE THIS TO A EXFIL SENARIO LATER
)

//END ONLY CHANGE THESE -------

// Create and parse the command line args
func (cmd *CommandArgs) Parse() {

	// Parse all of the flags
	flag.Parse()

	// CHECK REQUIRED ARGS HERE
	// Make sure you add the run of the function below to test if its required
	_ = cmd.ListProtocols()
	_ = cmd.ListExfils()
	_ = cmd.getInterface()
	_ = cmd.getNextHop()
	_ = cmd.getRemoteProtocol()
	_ = cmd.GetExfilProtocol()

}

// Checks to see if we are using this as an endpoint
func (cmd *CommandArgs) IsEndPoint() bool {
	if *ExfilProtocol == "" {
		return false
	}

	return true
}

// Checks for the Exfil Protocol Option
func (cmd *CommandArgs) GetExfilProtocol() string {

	if *ExfilProtocol == "" {
		return ""
	}

	if containsStr(avilExfil, strings.ToUpper(*ExfilProtocol)) {
		return strings.ToUpper(*ExfilProtocol)
	} else {
		logData("Exfil Must be part of the allowed selections! Please see -le for available options.", true, true)
		os.Exit(1)
	}

	return ""
}

// List Exfils and Exit
func (cmd *CommandArgs) ListExfils() bool {
	if *ListExfil {
		logData("Available Exfil List:", true, false)
		for _, prot := range avilExfil {
			fmt.Println("\t", prot)
		}
		fmt.Println()
		os.Exit(0)
	}
	return false
}

// List Protocols and Exit
func (cmd *CommandArgs) ListProtocols() bool {
	if *ListProts {
		logData("Available Protocol List:", true, false)
		for _, prot := range avilProtocols {
			fmt.Println("\t", prot)
		}
		fmt.Println()
		os.Exit(0)
	}
	return false
}

// Check the File input Flag
func (cmd *CommandArgs) isSetExfilFile() bool {

	if *ExfilFile != "" {
		return true
	}
	return false

}

// Handle the File input Flag
func (cmd *CommandArgs) GetExfilFile() string {
	return *ExfilFile
}

// Handle the Dynamically Ran Flag
func (cmd *CommandArgs) isDynamicRun() bool {
	if *DynamicRun && *Port == "" {
		return true
	}
	return false
}

// Handle the interface Flag
func (cmd *CommandArgs) getInterface() string {
	if *NetInterface == "" {
		logData("-i Required & Should be a Valid Interface! Use -h for more information..", true, true)
		os.Exit(0)
	}

	return *NetInterface
}

// Handle the Specified Protocol Flag
func (cmd *CommandArgs) getRemoteProtocol() string {
	if *DynamicRun && *Port != "" && *ExfilProtocol == "" {
		logData("Cannot use a custom protocol with Dynamic setting enabled (-d)", true, true)
		os.Exit(1)
	}
	if *DynamicRun == false && *Port == "" && *ExfilProtocol == "" {
		logData("You must either use dynamic (-d) or set a protocol (-protocol HTTP)!", true, true)
		os.Exit(1)
	}

	if containsStr(avilProtocols, strings.ToUpper(*Port)) {
		return strings.ToUpper(*Port)
	}

	return "HTTP"
}

// Handle the interface Flag
func (cmd *CommandArgs) getNextHop() string {
	// Handle if it was set to default or local host - fail if this is the case
	host := *NextHop

	if host == "127.0.0.1" && *ExfilProtocol == "" {
		logData("-h Required & Cannot be 127.0.0.1! Use -h for more information..", true, true)
		os.Exit(0)
	}

	return host
}

func printBanner() {
	banner := `
                                                           *,.......**                             
                                                         /.       .   ,,                           
                                                        .,. ...   ..  ./                           
                                                        .,..  ...  .. .*                           
                                                        .,. ..... ..   ,.                          
                                                        .,   ,...,.,.  ,.                          
                                                        .,...,.,. ...  ,,                          
                                                         * . ,*,*....  ,,                          
                                                         *. ... . ..   .,                          
                                                         *...  .    .  .,                          
                                                         *.   ..   ...,,#%#*,,/%,                  
                                                      .*(*.((//*.****..,*,,,,,,,,,*&/              
                                          .,(%%%(*....,,&(.,*/*,.****.*%*#,,,,,,,,,,,,##.          
                                   ,&(,..,,*(%#((%*,,,,*/,,..,*//*,,..,,(/,,,,,,,,,,,,,,,,(%.      
                                  &..,,,,,,,#(((((#(#,,,,(%#/*,,,**/#%(,,,,,,,,,........,//(*&     
                                .(((&*,,,,,/#(((((#(%,,,,,,,,,,,,,,.......,*/(#(***********%     
                               #(((((((%*,,,,,,,(*,,,,,,........./(#((***********************%     
                               &(((((%%&,,,,,,,......,*(#((**********************************%     
                             %*,/&%%%&(,,,...(#/**************************************/(%%(*.      
                          /%,,,,,,,,,,,,...(*********************************/(%%(*,.,,,.          
                       #(,,,,,,,,,......../*************************/#%%/,.   .       .            
                  ,&/,,,,,,,,........,,,,..(*************(#/.            .                       
              *%(,,,,,,,,.......*%(.       (%(//(%%#*.           ..                                
         /%/,,,,,,,,,......,,&*           .           .                                            
      /#,,,,,,,,,,......,/%,                                                                       
     #*,,,,,,,,......,*&.                                                                          
     %...,,,,.....,/%.                                                                             
     #.........,,%/                                                                                
	  /(....,/%.       
		     
		$$$$$$\    $$\   $$$$$$\  $$$$$$$$\ $$$$$$$$\ $$\   $$\ $$$$$$$$\ $$$$$$\ $$\       
		$$  __$$\ $$$$ | $$ ___$$\ \____$$  |$$  _____|$$ |  $$ |$$  _____|\_$$  _|$$ |      
		$$ /  \__|\_$$ | \_/   $$ |    $$  / $$ |      \$$\ $$  |$$ |        $$ |  $$ |      
		$$ |        $$ |   $$$$$ /    $$  /  $$$$$\     \$$$$  / $$$$$\      $$ |  $$ |      
		$$ |        $$ |   \___$$\   $$  /   $$  __|    $$  $$<  $$  __|     $$ |  $$ |      
		$$ |  $$\   $$ | $$\   $$ | $$  /    $$ |      $$  /\$$\ $$ |        $$ |  $$ |      
		\$$$$$$  |$$$$$$\\$$$$$$  |$$  /     $$$$$$$$\ $$ /  $$ |$$ |      $$$$$$\ $$$$$$$$\ 
		 \______/ \______|\______/ \__/      \________|\__|  \__|\__|      \______|\________|																																																								  
			`
	fmt.Print(banner)
	fmt.Print("\t\tLets's Exfil Morty\n")
	fmt.Println("\n\t\tAuthor: Brandon Dennis\n\n")

}
