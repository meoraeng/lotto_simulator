package httpapi

import "net/http"

// HTTP 핸들러 등록을 위한 공통 인터페이스
type RouteRegistrar interface {
	Register(mux *http.ServeMux)
}
