package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

type Cotacao struct {
	USD struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "./cotacoes.db")
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}
	defer db.Close()

	if err := inicializarBanco(); err != nil {
		log.Fatal("Erro ao inicializar banco:", err)
	}

	http.HandleFunc("/cotacao", cotacaoHandler)
	log.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func inicializarBanco() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, bid TEXT, data TIMESTAMP)`)
	return err
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctxAPI, cancelAPI := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancelAPI()

	req, err := http.NewRequestWithContext(ctxAPI, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		http.Error(w, "Erro ao criar requisição externa", http.StatusInternalServerError)
		log.Println("Erro criar requisição externa:", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Erro ao obter cotação", http.StatusGatewayTimeout)
		log.Println("Timeout API externa:", err)
		return
	}
	defer resp.Body.Close()

	var c Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		http.Error(w, "Erro ao decodificar resposta", http.StatusInternalServerError)
		log.Println("Erro ao decodificar JSON:", err)
		return
	}

	if c.USD.Bid == "" {
		log.Println("Valor 'bid' vazio, não será salvo no banco.")
		http.Error(w, "Cotação inválida", http.StatusInternalServerError)
		return
	}

	go salvarCotacao(c.USD.Bid) // salva de forma assíncrona com contexto

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"bid": c.USD.Bid,
	})
}

func salvarCotacao(bid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, `INSERT INTO cotacoes(bid, data) VALUES (?, ?)`, bid, time.Now())
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout ao tentar salvar no banco (10ms)")
		} else {
			log.Println("Erro ao inserir cotação:", err)
		}
		return
	}

	log.Println("Cotação salva no banco com sucesso!")
}
