package httpapi

import (
	"net/http"
)

// 메시지 남기는 경우
func writeErrorMsg(w http.ResponseWriter, status int, msg string) {
	http.Error(w, "[ERROR] "+msg, status)
}

// 에러 객체 함께 남기는 경우
func writeError(w http.ResponseWriter, status int, msg string, err error) {
	if err != nil {
		http.Error(w, "[ERROR] "+msg+": "+err.Error(), status)
		return
	}
	http.Error(w, "[ERROR] "+msg, status)
}
