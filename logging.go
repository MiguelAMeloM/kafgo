package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// logRequestHandler es un middleware que imprime la información de la request y el status code.
func logRequestHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Envolver el ResponseWriter para capturar el status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		// Estilo tipo GinGonic
		method := padRight(r.Method, 6)
		path := r.URL.Path
		status := lrw.statusCode
		statusColor := colorForStatus(status)
		methodColor := colorForMethod(r.Method)

		log.Printf("%s |%s %3d %s| %s | %s | %s\n",
			start.Format("2006/01/02 - 15:04:05"),
			statusColor, status, resetColor(),
			duration,
			methodColor+method+resetColor(),
			path,
		)
	}
}

// loggingResponseWriter captura el status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader intercepta el código de estado
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Helpers para colorear la salida (solo si se usa terminal compatible ANSI)
const (
	green  = "\033[32m"
	white  = "\033[97m"
	yellow = "\033[33m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	reset  = "\033[0m"
)

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return cyan
	case "POST":
		return green
	case "PUT":
		return yellow
	case "DELETE":
		return red
	default:
		return white
	}
}

func resetColor() string {
	return reset
}

func padRight(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return str + strings.Repeat(" ", length-len(str))
}
