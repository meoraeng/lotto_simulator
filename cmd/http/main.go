package main

import (
	"log"
	"net/http"

	"github.com/meoraeng/lotto_simulator/internal/httpapi"
)

const httpAddr = ":8080"

func main() {
	mux := http.NewServeMux() // 라우터

	h := httpapi.NewHandler()
	h.Register(mux) // 핸들러가 엔드포인트들을 라우터에 등록

	log.Printf("HTTP server listening on %s\n", httpAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal(err)
	}
}
