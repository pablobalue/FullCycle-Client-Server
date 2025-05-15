package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal("Erro criando requisição:", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Timeout ou erro na requisição:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Erro lendo resposta:", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal("Erro decodificando JSON:", err)
	}

	bid := result["bid"]
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatal("Erro criando arquivo:", err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s\n", bid))
	if err != nil {
		log.Fatal("Erro escrevendo no arquivo:", err)
	}

	log.Println("Cotação salva em cotacao.txt:", bid)
}
