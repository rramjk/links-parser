package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

// test

func main() {
	startProgram()
}

func startProgram() {
	startTime := time.Now()
	var requestSource string
	var directoryForParse string
	parseParam(requestSource, directoryForParse)

	fmt.Println(requestSource + " " + directoryForParse)

	fmt.Printf(" Время работы программы: %v\n", time.Now().Sub(startTime))
}

func parseParam(src string, dst string) {
	flag.StringVar(&src, "src", "null", "address for request")
	flag.StringVar(&dst, "dst", "null", "directory for response")
	flag.Parse()
	if src == "null" || dst == "null" {
		fmt.Println(errors.New("Недостаточно параметров!"))
		os.Exit(0)
	}
}
