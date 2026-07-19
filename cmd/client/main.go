package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	endpoint := "http://localhost:8080/"

	fmt.Println("Введите длинный URL:")

	reader := bufio.NewReader(os.Stdin)

	longURL, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	longURL = strings.TrimSpace(longURL)

	request, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		strings.NewReader(longURL),
	)
	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Статус:", response.Status)
	fmt.Println("Короткая ссылка:", string(body))
}
