package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/meoraeng/lotto_simulator/internal/httpapi"
	"github.com/meoraeng/lotto_simulator/internal/webui"
)

func main() {
	mux := http.NewServeMux()

	// 여러 핸들러 타입을 공통 인터페이스로 처리
	registrars := []httpapi.RouteRegistrar{
		mustNewWebUIHandler(),
		httpapi.NewHandler(),
	}

	for _, registrar := range registrars {
		registrar.Register(mux)
	}

	log.Println("서버 실행중:  http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func mustNewWebUIHandler() *webui.Handler {
	templatesPath := filepath.Join("internal", "webui", "templates")
	h, err := webui.NewHandler(templatesPath)
	if err != nil {
		log.Fatal(err)
	}
	return h
}
