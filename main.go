package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// test

func main() {
	startParse()
}

func startParse() {
	startTime := time.Now()
	var requestSource string
	var directoryForParse string

	parseParam(&requestSource, &directoryForParse)

	fileRead, errRead := os.Open(requestSource)
	throwException(errRead)

	parseLinks(fileRead, directoryForParse)

	fmt.Printf("Время работы программы: %v\n", time.Now().Sub(startTime))
}

func parseParam(src *string, dst *string) {
	flag.StringVar(src, "src", "null", "address for request")
	flag.StringVar(dst, "dst", "null", "directory for response")
	flag.Parse()
	if *src == "null" || *dst == "null" {
		throwLocalException()
	}
}

func throwLocalException() {
	fmt.Println(errors.New("Ошибка работы программы!"))
	os.Exit(0)
}

func throwException(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

}

func parseLinks(file io.Reader, directory string) {
	reader := bufio.NewReader(file)
	for {
		link, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		throwException(err)
		sendResponse(link, directory)
	}
}

func sendResponse(link string, directory string) {
	link = validateLink(link)
	if link == "" {
		return
	}
	request, err := http.NewRequest("GET", "http://"+link, nil)
	throwException(err)
	client := &http.Client{}
	response, err := client.Do(request)
	throwException(err)
	body, err := ioutil.ReadAll(response.Body)
	throwException(err)

	fileWrite, err := os.Create(directory + "/" + link)
	throwException(err)

	_, err = fileWrite.WriteString(string(body))
	throwException(err)

	fmt.Println("ready!")
}

func validateLink(link string) string {
	if link[len(link)-1:] == "\n" {
		return link[:len(link)-1]
	}
	return link
}
