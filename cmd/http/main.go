package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/meoraeng/lotto_simulator/internal/webui"
)

func main() {
	mux := http.NewServeMux()

	// 템플릿 경로
	templatesPath := filepath.Join("internal", "webui", "templates")

	// web UI 핸들러
	h, err := webui.NewHandler(templatesPath)
	if err != nil {
		log.Fatal(err)
	}
	h.Register(mux)

	log.Println("서버 실행중:  http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
