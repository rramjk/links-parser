package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// получаем время начала программы
	startTime := time.Now()

	// путь для предполагаемого файла и дирректории
	var requestSource string
	var directoryForParse string
	err := parseParam(&requestSource, &directoryForParse)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	// открытие файла
	fileRead, errRead := os.Open(requestSource)
	if errRead != nil {
		fmt.Println("Ошибка открытия файла!")
	}
	defer fileRead.Close()
	//считывание ссылок из файла
	parseLinks(fileRead, directoryForParse)

	fmt.Printf("Время работы программы: %v\n", time.Now().Sub(startTime))
}

// parseParam - получение параметров с вызова программы
func parseParam(src *string, dst *string) error {
	flag.StringVar(src, "src", "null", "address for request")
	flag.StringVar(dst, "dst", "null", "directory for response")
	flag.Parse()

	// проверка на корректность полученных параметров
	if !srcAndDstIsCorrect(*src, *dst) {
		return errors.New("Параметр src должен содержать путь к файлу, а dst путь к папке финального назначения")
	}
	return nil
}

// srcAndDstIsCorrect - проверка источника и необходимой дирректории на корректность (true - src: удачный путь к файлу dst: корректная папка)
func srcAndDstIsCorrect(src string, dst string) bool {
	if src == "null" || dst == "null" {
		fmt.Print(errors.New("Источник или директория указаны не верно\n"))
		fmt.Println("./main --src='path to file' --dst='path to directory'")
		os.Exit(0)
	}
	if len(src) < 3 || len(dst) < 3 {
		fmt.Println(errors.New("Параметр src должен содержать путь к файлу, а dst путь к папке финального назначения"))
		os.Exit(0)
	}
	// тестовое получение папки если оно удачно значит создавать новую папку не стоит
	_, err := os.Stat(dst)
	if dst[:2] == "./" && err != nil {
		err := os.Mkdir(dst[2:], os.ModePerm)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
	srcInfo, srcErr := os.Stat(src)
	dstInfo, dstErr := os.Stat(dst)

	if srcErr != nil || dstErr != nil {
		if os.IsNotExist(srcErr) || os.IsNotExist(dstErr) {
			fmt.Println("Файла или директория не существует")
		} else {
			fmt.Println("Ошибка получения информации о файле или директории")
		}
		if srcErr != nil {
			fmt.Println(srcErr)
			os.Exit(0)
		} else if dstErr != nil {
			fmt.Println(srcErr)
			os.Exit(0)
		}
	}
	return dstInfo.IsDir() && !srcInfo.IsDir()
}

// parseLinks - считывание в файле ссылки построчно (\n Enter) и записывание файл в дирректорию
func parseLinks(file io.Reader, directory string) {
	reader := bufio.NewReader(file)
	for {
		link, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		err = writeResponseBody(link, directory)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// writeResponseBody - запись ссылки в созданный по директории файл, в случае ошибки выходит из функции
func writeResponseBody(link string, directory string) error {
	link = validateLink(link)

	request, err := http.NewRequest("GET", fmt.Sprintf("http://%v", link), nil)
	if err != nil {
		err = errors.New(fmt.Sprintf("Ошибка при создании запроса: http://%v", link))
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		err = errors.New(fmt.Sprintf("Ошибка при отправке запроса: http://%v", link))
		return err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		err = errors.New(fmt.Sprintf("Ошибка при чтении ответа: http://%v", link))
		return err
	}

	err = writeBodyInDirectory(getDomainInLink(*request), directory, string(body))
	if err != nil {
		return err
	} else {
		fmt.Println(fmt.Sprintf("http://%v - успешно записан", link))
	}
	return nil
}

// writeBodyInDirectory - создание и запись информации в файл по указанной директории
func writeBodyInDirectory(fileName string, directoryPath string, textForFile string) error {
	fileWrite, err := os.Create(fmt.Sprintf("%v/%v", directoryPath, fileName))
	if err != nil {
		return errors.New("Ошибка создания файла!")
	}

	_, err = fileWrite.WriteString(textForFile)
	if err != nil {
		return errors.New("Ошибка записи в файл!")
	}

	err = fileWrite.Close()
	if err != nil {
		return err
	}
	return nil
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
