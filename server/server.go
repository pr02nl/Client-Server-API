package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	//sqlite3 driver
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Usdbrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}
type CotacaoApi struct {
	Usdbrl Usdbrl `json:"USD"`
}

type OutPutDTO struct {
	Bid float64 `json:"bid"`
}

type Server struct {
	db *sql.DB
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	usdbrl, err := getCotacao()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = saveCotacaoDb(s.db, usdbrl)
	if err != nil {
		log.Panicln("saveCotacaoDb", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	bid, err := strconv.ParseFloat(usdbrl.Bid, 64)
	if err != nil {
		log.Panicln("strconv.ParseFloat", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	outPut := OutPutDTO{Bid: bid}
	json.NewEncoder(w).Encode(outPut)
}

func main() {
	db, err := sql.Open("sqlite3", "cotacao.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	s := Server{db: db}
	mux := http.NewServeMux()
	mux.Handle("/cotacao", &s)
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func getCotacao() (*Usdbrl, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/all/USD-BRL", nil)
	if err != nil {
		log.Println("NewRequestWithContext", err)
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("DefaultClient.Do", err)
		return nil, err
	}
	defer res.Body.Close()
	var cotacaoApi CotacaoApi
	err = json.NewDecoder(res.Body).Decode(&cotacaoApi)
	if err != nil {
		log.Println("NewDecoder", err)
		return nil, err
	}
	fmt.Println(cotacaoApi)
	return &cotacaoApi.Usdbrl, nil
}

func saveCotacaoDb(db *sql.DB, usdbrl *Usdbrl) error {
	stmt, err := db.Prepare("insert into cotacao(id, bid, create_date) values(?, ?, ?)")
	if err != nil {
		log.Println("db.Prepare", err)
		return err
	}
	defer stmt.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	bid, err := strconv.ParseFloat(usdbrl.Bid, 64)
	if err != nil {
		log.Println("strconv.ParseFloat", err)
		return err
	}
	createDate, err := time.Parse("2006-01-02 15:04:05", usdbrl.CreateDate)
	if err != nil {
		log.Println("time.Parse", err)
		return err
	}
	_, err = stmt.ExecContext(ctx, uuid.New(), bid, createDate)
	if err != nil {
		log.Println("stmt.ExecContext", err)
		return err
	}
	return nil
}
