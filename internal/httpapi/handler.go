package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// 핸들러 함수 연결
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/round", h.handleCalculateRound)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func (h *Handler) handleCalculateRound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// post가 아닌 경우
		writeErrorMsg(w, http.StatusMethodNotAllowed, "허용되지 않은 메서드입니다")
		return
	}

	var in lotto.RoundInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil { // HTTP body를 RoundInput으로 디코딩
		writeError(w, http.StatusBadRequest, "유효하지 않은 JSON입니다", err)
		return
	}

	out, err := lotto.CalculateRound(in) // 도메인 로직 호출

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		// 도메인에서 넘어온 에러 종류에 따라 HTTP 상태코드 및 메시지 매핑
		if errors.Is(err, lotto.ErrInvalidMode) {
			writeErrorMsg(w, http.StatusBadRequest, "잘못된 모드 값입니다")
			return
		}

		// 예상 못한 도메인 에러
		writeError(w, http.StatusInternalServerError, "서버 내부 오류가 발생했습니다", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		writeError(w, http.StatusInternalServerError, "결과 인코딩에 실패했습니다", err)
		return
	}
}
