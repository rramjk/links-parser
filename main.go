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
	"time"
)

func main() {
	startParse()
}

// startParse - старт работы парсера
func startParse() {
	// получаем время начала программы
	startTime := time.Now()

	// путь для предполагаемого файла и дирректории
	var requestSource string
	var directoryForParse string
	parseParam(&requestSource, &directoryForParse)
	// открытие файла
	fileRead, errRead := os.Open(requestSource)
	exceptionIsNotNIL(errRead)
	//считывание ссылок из файла
	parseLinks(fileRead, directoryForParse)

	fmt.Printf("Время работы программы: %v\n", time.Now().Sub(startTime))
}

// parseParam - получение параметров с вызова программы
func parseParam(src *string, dst *string) {
	flag.StringVar(src, "src", "null", "address for request")
	flag.StringVar(dst, "dst", "null", "directory for response")
	flag.Parse()
	// проверка на корректность полученных параметров
	if !srcAndDstIsCorrect(*src, *dst) {
		exceptionIsNotNIL(errors.New("Параметр src содержит файл, а dst папку финального назначения"))
	}
}

// srcAndDstIsCorrect - проверка источника и необходимой дирректории на корректность (true - src: удачный путь к файлу dst: корректная папка)
func srcAndDstIsCorrect(src string, dst string) bool {
	if src == "null" || dst == "null" {
		exceptionIsNotNIL(errors.New("Источник или дирректория не найдены"))
	}
	// тестовое получение папки если оно удачно значит создавать новую папку не стоит
	_, err := os.Stat(dst)
	if dst[:2] == "./" && err != nil {
		err := os.Mkdir(dst[2:], os.ModePerm)
		exceptionIsNotNIL(err)
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

// exceptionIsNotNIL - прекращение работы программы в случае ошибки
func exceptionIsNotNIL(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

}

// parseLinks - считывание в файле ссылки построчно (\n Enter) и записывание файл в дирректорию
func parseLinks(file io.Reader, directory string) {
	reader := bufio.NewReader(file)
	for {
		link, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		exceptionIsNotNIL(err)
		writeResponseBody(link, directory)
	}
}

// writeResponseBody - запись ссылки в созданный по директории файл, в случае ошибки выходит из функции
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

// writeBodyInDirectory - создание и запись информации в файл по указанной директории
func writeBodyInDirectory(fileName string, directoryPath string, textForFile string) {
	fileWrite, err := os.Create(directoryPath + "/" + fileName)
	exceptionIsNotNIL(err)

	_, err = fileWrite.WriteString(textForFile)
	exceptionIsNotNIL(err)
}

// validateLink - валидация и коррекция ссылки на лишнюю информацию
func validateLink(link string) string {
	link = strings.ReplaceAll(link, "\n", "")
	link = strings.ReplaceAll(link, "http://", "")
	link = strings.ReplaceAll(link, "https://", "")
	return link
}

// getDomainInLink - получение домена из запроса
func getDomainInLink(req http.Request) string {
	return string(req.Host)
}
