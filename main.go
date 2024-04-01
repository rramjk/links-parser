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

var noSuchDirectoryError = errors.New("Директории не существует")

func main() {
	// получаем время начала программы
	startTime := time.Now()

	// путь для предполагаемого файла и дирректории
	var requestSource string
	var directoryForParse string
	err := parseParam(&requestSource, &directoryForParse)
	if err == noSuchDirectoryError {
		err = createDrtInCurrentFolder(directoryForParse)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// открытие файла
	fileRead, errRead := os.Open(requestSource)
	if errRead != nil {
		fmt.Println("Ошибка открытия файла!")
	}
	defer fileRead.Close()
	//считывание ссылок из файла
	errParse := readLinksAndCreateFiles(fileRead, directoryForParse)
	if errParse != nil {
		fmt.Println(errParse)
		os.Exit(1)
	}
	fmt.Printf("Время работы программы: %v\n", time.Now().Sub(startTime))
}

// parseParam - получение параметров с вызова программы
func parseParam(src *string, dst *string) error {
	flag.StringVar(src, "src", "null", "address for request")
	flag.StringVar(dst, "dst", "null", "directory for response")
	flag.Parse()

	// проверка на корректность полученных параметров
	err := srcAndDstIsCorrect(*src, *dst)
	if err != nil {
		return err
	}

	return nil
}

// srcAndDstIsCorrect - проверка источника и необходимой дирректории на корректность (true - src: удачный путь к файлу dst: корректная папка)
func srcAndDstIsCorrect(src string, dst string) error {
	if src == "null" || dst == "null" || len(src) < 2 || len(dst) < 2 {
		return errors.New("Источник или директория указаны не верно\n./main --src='path to file' --dst='path to directory'")
	}
	// тестовое получение папки если оно удачно значит создавать новую папку не стоит
	_, err := os.Stat(dst)
	if dst[:2] == "./" && err != nil {
		return noSuchDirectoryError
	}
	srcInfo, srcErr := os.Stat(src)
	dstInfo, dstErr := os.Stat(dst)

	if srcErr != nil || dstErr != nil {
		if os.IsNotExist(srcErr) || os.IsNotExist(dstErr) {
			return errors.New("Файл или директория не существует")
		} else {
			return errors.New("Ошибка получения информации о файле или директории")
		}
	}
	if !(dstInfo.IsDir() && !srcInfo.IsDir()) {
		return errors.New("Параметр src должен содержать путь к файлу, а dst путь к папке финального назначения")
	}
	return nil
}
func createDrtInCurrentFolder(dstPath string) error {
	err := os.Mkdir(dstPath[2:], os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// readLinksAndCreateFiles - считывание в файле ссылки построчно (\n Enter) и записывание файл в дирректорию
func readLinksAndCreateFiles(file io.Reader, directory string) error {
	reader := bufio.NewReader(file)
	for {
		link, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = writeResponseBody(link, directory)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
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
	defer response.Body.Close()

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
	defer fileWrite.Close()
	_, err = fileWrite.WriteString(textForFile)
	if err != nil {
		return errors.New("Ошибка записи в файл!")
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
