package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"pos-toko/go-print/internal/api"
	"pos-toko/go-print/internal/printer"
	"pos-toko/go-print/internal/receipt"
)

const (
	defaultEnvFile = ".env"
	defaultAddr    = ":8080"
	defaultDevice  = "/dev/cu.RPP02N"
)

func main() {
	if err := loadEnv(defaultEnvFile); err != nil && !os.IsNotExist(err) {
		log.Printf("could not load %s: %v", defaultEnvFile, err)
	}

	addr := getenv("HTTP_ADDR", defaultAddr)
	devicePath := getenv("PRINTER_DEVICE", defaultDevice)
	baudRate := getenv("BAUD_RATE", "9600")
	receiptMode := receipt.Mode(getenv("RECEIPT_MODE", string(receipt.ModeESCPOS)))

	writer := printer.NewSerialWriter(devicePath, baudRate, 250*time.Millisecond)
	service := receipt.NewService(receiptMode)
	handler := api.NewHandler(service, writer)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("printer service listening on %s, device=%s, baud_rate=%s, receipt_mode=%s", addr, devicePath, baudRate, receiptMode)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
