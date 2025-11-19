package webui

import "net/http"

const errorPrefix = "[ERROR] "

func errorText(msg string) string {
	return errorPrefix + msg
}

func errorMsg(err error) string {
	if err == nil {
		return ""
	}
	return errorPrefix + err.Error()
}

// 그냥 HTTP Response에 에러 텍스트만 내려야 할 때
func writeHttpErrorText(w http.ResponseWriter, status int, msg string) {
	http.Error(w, errorText(msg), status)
}

// 에러 객체까지 함께 찍고 싶을 때
func writeHttpError(w http.ResponseWriter, status int, msg string, err error) {
	if err != nil {
		http.Error(w, errorText(msg)+": "+err.Error(), status)
		return
	}
	http.Error(w, errorText(msg), status)
}
