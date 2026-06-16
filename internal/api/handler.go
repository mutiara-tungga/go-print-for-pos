package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"pos-toko/go-print/internal/printer"
	"pos-toko/go-print/internal/receipt"
)

type ReceiptService interface {
	Build(receipt.Request) ([]byte, error)
}

type Printer interface {
	Write([]byte) error
}

type Handler struct {
	receipts ReceiptService
	printer  Printer
}

func NewHandler(receipts ReceiptService, printer Printer) *Handler {
	return &Handler{receipts: receipts, printer: printer}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /print/receipt", h.printReceipt)
	mux.HandleFunc("POST /print/test", h.printTest)
}

func (h *Handler) printReceipt(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req receipt.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON payload: %v", err))
		return
	}

	data, err := h.receipts.Build(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.printer.Write(data); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, printer.ErrUnavailable) {
			status = http.StatusServiceUnavailable
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "printed"})
}

func (h *Handler) printTest(w http.ResponseWriter, r *http.Request) {
	data := []byte("HELLO PUTRI\nTEST PRINT\n----------------\n")
	if err := h.printer.Write(data); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, printer.ErrUnavailable) {
			status = http.StatusServiceUnavailable
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "printed"})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
