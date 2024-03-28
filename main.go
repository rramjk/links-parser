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
	"strings"
	"sync"
	"time"
)

// test

func main() {
	startParse()
}

/*
доделать валидацию и корректность введены данных
*/
func startParse() {
	// получаем время начала программы
	startTime := time.Now()

	// путь для предполагаемого файла и дирректории
	var requestSource string
	var directoryForParse string
	parseParam(&requestSource, &directoryForParse)

	fileRead, errRead := os.Open(requestSource)
	exceptionIsNotNIL(errRead)

	parseLinks(fileRead, directoryForParse)

	fmt.Printf("Время работы программы: %v\n", time.Now().Sub(startTime))
}

// получаем параметры с вызова программы
func parseParam(src *string, dst *string) {
	flag.StringVar(src, "src", "null", "address for request")
	flag.StringVar(dst, "dst", "null", "directory for response")
	flag.Parse()

	if !srcAndDstIsCorrect(*src, *dst) {
		exceptionIsNotNIL(errors.New("Параметр src содержит файл, а dst папку финального назначения"))
	}
}
func srcAndDstIsCorrect(src string, dst string) bool {
	if src == "null" || dst == "null" {
		exceptionIsNotNIL(errors.New("Источник или дирректория не найдены"))
	}
	srcInfo, srcErr := os.Stat(src)
	dstInfo, dstErr := os.Stat(dst)

	if srcErr != nil || dstErr != nil {
		if os.IsNotExist(srcErr) || os.IsNotExist(dstErr) {
			fmt.Println("Файл или директория не существует")
		} else {
			fmt.Println("Ошибка получения информации о файле или директории")
		}
		exceptionIsNotNIL(srcErr)
		exceptionIsNotNIL(dstErr)
	}
	return dstInfo.IsDir() && !srcInfo.IsDir()
}

func exceptionIsNotNIL(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

}

func parseLinks(file io.Reader, directory string) {
	reader := bufio.NewReader(file)
	var wgroup sync.WaitGroup
	for {
		link, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		exceptionIsNotNIL(err)
		wgroup.Add(1)
		go func(link string, directory string) {
			defer wgroup.Done()
			writeResponseBody(link, directory)
		}(link, directory)
		wgroup.Wait()

	}
}

func writeResponseBody(link string, directory string) {
	link = validateLink(link)

	request, err := http.NewRequest("GET", "http://"+link, nil)
	if err != nil {
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	writeBodyInDirectory(getDomainInLink(*request), directory, string(body))
}

func writeBodyInDirectory(fileName string, directoryPath string, textForFile string) {
	fileWrite, err := os.Create(directoryPath + "/" + fileName)
	exceptionIsNotNIL(err)

	_, err = fileWrite.WriteString(textForFile)
	exceptionIsNotNIL(err)
}
func validateLink(link string) string {
	link = strings.ReplaceAll(link, "\n", "")
	link = strings.ReplaceAll(link, "http://", "")
	link = strings.ReplaceAll(link, "https://", "")
	return link
}
func getDomainInLink(req http.Request) string {
	return string(req.Host)
}
