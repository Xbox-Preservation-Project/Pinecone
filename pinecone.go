package main

import (
	"flag"
	"fmt"
)

var (
	titles        TitleList
	updateFlag    = false
	summarizeFlag = false
	titleIDFlag   = ""
	fatxplorer    = false
	dumpLocation  = "dump"
	helpFlag      = false
	version       = "0.5.1"
)

func main() {
	flag.BoolVar(&updateFlag, "update", false, "Update the JSON data from the source URL")
	flag.BoolVar(&updateFlag, "u", false, "Update the JSON data from the source URL")
	flag.BoolVar(&summarizeFlag, "summarize", false, "Print summary statistics for all titles")
	flag.BoolVar(&summarizeFlag, "s", false, "Print summary statistics for all titles")
	flag.StringVar(&titleIDFlag, "titleid", "", "Filter statistics by Title ID")
	flag.StringVar(&titleIDFlag, "tID", "", "Filter statistics by Title ID")
	flag.BoolVar(&fatxplorer, "fatxplorer", false, "Use FatXplorer's X: drive")
	flag.BoolVar(&fatxplorer, "f", false, "Use FatXplorer's X: drive")
	flag.StringVar(&dumpLocation, "location", "dump", "Directory to search for TDATA/UDATA directories")
	flag.StringVar(&dumpLocation, "l", "dump", "Directory to search for TDATA/UDATA directories")
	flag.BoolVar(&helpFlag, "help", false, "Display help information")
	flag.BoolVar(&helpFlag, "h", false, "Display help information")

	flag.Parse() // Parse command line flags

	// Check for help flag
	if helpFlag {
		fmt.Println("Usage of Pinecone:")
		fmt.Println("  -u, --update:     Update the JSON data from the source URL. If not set, uses local copies of data.")
		fmt.Println("  -s, --summarize:  Print summary statistics for all titles. If not set, checks for content in the TDATA folder.")
		fmt.Println("  -tID, --titleid:  Filter statistics by Title ID (-titleID=ABCD1234). If not set, statistics are computed for all titles.")
		fmt.Println("  -f, --fatxplorer: Use FATXPlorer's X drive as the root directory. If not set, runs as normal. (Windows Only)")
		fmt.Println("  -l --location:    Directory where TDATA/UDATA folders are stored. If not set, checks in \"dump\"")
		fmt.Println("  -h, --help:       Display this help information.")
		return
	}

	jsonFilePath := "data/id_database.json"
	jsonDataFolder := "data"
	jsonURL := "https://api.github.com/repos/Xbox-Preservation-Project/Pinecone/contents/data/id_database.json"

	cliOpts := CLIOptions{
		DataFolder:   jsonDataFolder,
		JSONFilePath: jsonFilePath,
		JSONUrl:      jsonURL,
	}

	startCLI(cliOpts)
}
