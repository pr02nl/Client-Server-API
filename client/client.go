package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type InputDTO struct {
	Bid float64 `json:"bid"`
}

func main() {
	f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	input, err := getCotacao()
	if err != nil {
		panic(err)
	}
	err = saveCotacaoFile(f, input)
	if err != nil {
		panic(err)
	}
}

func getCotacao() (*InputDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}
	defer res.Body.Close()
	var input InputDTO
	err = json.NewDecoder(res.Body).Decode(&input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func saveCotacaoFile(f *os.File, input *InputDTO) error {
	n, err := fmt.Fprintln(f, "Dólar: ", input.Bid)
	if err != nil {
		return err
	}
	fmt.Println("Foi gravado ", n, " bytes no arquivo")
	return nil
}
