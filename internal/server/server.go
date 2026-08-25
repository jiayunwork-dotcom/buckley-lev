package server

import "net/http"

func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/profile", handleProfile)
	mux.HandleFunc("/api/fractional", handleFractional)
	mux.HandleFunc("/api/sweep", handleSweep)
	mux.HandleFunc("/api/history", handleHistory)
	return mux
}

func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, New())
}
