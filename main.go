package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

const OPENAI_API_KEY = "your_openai_api_key"

func getCompletion(userPrompt string, contexts []string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	systemPrompt := fmt.Sprintf(`
	Kamu adalah asisten yang membantu.
	Gunakan konteks berikut ini untuk menjawab pertanyaan.
	Jika Kamu tidak tahu jawabannya, katakan bahwa Kamu tidak tahu.

	%s`, strings.Join(contexts, "\n"))

	payload, _ := json.Marshal(map[string]any{
		"model":       "gpt-4o-mini",
		"temperature": 0.7,
		"messages": []map[string]any{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	})

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", OPENAI_API_KEY))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("http request not returned ok: %s", string(body))
	}

	var r struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", err
	}

	if len(r.Choices) == 0 {
		return "", fmt.Errorf("no choices found in response")
	}

	return r.Choices[0].Message.Content, nil
}

func generateEmbeddings(input string) ([]float32, error) {
	url := "https://api.openai.com/v1/embeddings"
	payload, _ := json.Marshal(map[string]any{
		"input": input,
		"model": "text-embedding-3-small",
	})

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", OPENAI_API_KEY))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("http request not returned ok: %s", string(body))
	}

	var r struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	if len(r.Data) == 0 {
		return nil, fmt.Errorf("no data found in response")
	}

	return r.Data[0].Embedding, nil
}

func initDB() (*sql.DB, error) {
	// Replace with your actual connection details
	connStr := "postgresql://postgres:postgres@localhost:5432/ai_embedding?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func insertEmbedding(db *sql.DB, text string, embeddings []float32) error {
	query := "INSERT INTO documents (text, embedding) VALUES ($1, $2)"
	_, err := db.Exec(query, text, pgvector.NewVector(embeddings))
	if err != nil {
		return err
	}

	return nil
}

func similaritySearch(db *sql.DB, queryEmbeddings []float32) ([]string, error) {
	query := "SELECT text FROM documents ORDER BY embedding <-> $1 LIMIT 2"
	rows, err := db.Query(query, pgvector.NewVector(queryEmbeddings))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string

	if rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		results = append(results, text)
	}

	return results, nil
}

func main() {
	// Insert Embedding to Vector DB

	// input := `Pasangan Jenderal TNI (Purn) Prabowo Subianto dan Gibran Rakabuming Raka
	// secara resmi mengemban tugas sebagai Presiden dan Wakil Presiden Republik Indonesia
	// masa jabatan 2024-2029 pada Minggu, 20 Oktober 2024.
	// Keduanya dilantik dalam Sidang Paripurna Majelis Permusyawaratan Rakyat (MPR)
	// dalam rangka Pelantikan Presiden dan Wakil Presiden Masa Jabatan 2024-2029
	// yang diselenggarakan di Gedung Nusantara MPR/DPR/DPD RI, Jakarta.`

	// chunks := strings.Split(input, "/n/n")

	// for _, chunk := range chunks {
	// 	embedding, err := generateEmbeddings(chunk)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}

	// 	fmt.Println(embedding)
	// }

	query := "Siapkah presiden indonesia 2024?"
	embeddings, err := generateEmbeddings(query)
	if err != nil {
		panic(err.Error())
	}

	db, err := initDB()
	if err != nil {
		panic(err.Error())
	}

	contexts, err := similaritySearch(db, embeddings)
	if err != nil {
		panic(err.Error())
	}

	resp, err := getCompletion(query, contexts)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(resp)
}
