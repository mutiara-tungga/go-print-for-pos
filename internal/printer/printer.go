package printer

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var ErrUnavailable = errors.New("printer unavailable")

type SerialWriter struct {
	devicePath string
	retryDelay time.Duration
	closeDelay time.Duration
	mu         sync.Mutex
}

func NewSerialWriter(devicePath string, retryDelay time.Duration) *SerialWriter {
	return &SerialWriter{
		devicePath: devicePath,
		retryDelay: retryDelay,
		closeDelay: 150 * time.Millisecond,
	}
}

func (w *SerialWriter) Write(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(data) == 0 {
		return nil
	}

	log.Printf("printing %d bytes to %s", len(data), w.devicePath)
	err := w.writeOnce(data)
	if err == nil {
		log.Printf("printed %d bytes to %s", len(data), w.devicePath)
		return nil
	}

	if w.retryDelay > 0 {
		time.Sleep(w.retryDelay)
	}

	if retryErr := w.writeOnce(data); retryErr != nil {
		return fmt.Errorf("%w: write to %s failed after retry: %v; first error: %v", ErrUnavailable, w.devicePath, retryErr, err)
	}

	return nil
}

func (w *SerialWriter) writeOnce(data []byte) error {
	file, err := os.OpenFile(w.devicePath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("%w: open %s: %v", ErrUnavailable, w.devicePath, err)
	}
	defer file.Close()

	if err := writeAll(file, data); err != nil {
		return err
	}
	if w.closeDelay > 0 {
		time.Sleep(w.closeDelay)
	}
	return nil
}

func writeAll(file *os.File, data []byte) error {
	lines := bytes.SplitAfter(data, []byte("\n"))
	if len(lines) > 1 {
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			if err := writeChunk(file, line); err != nil {
				return err
			}
			time.Sleep(25 * time.Millisecond)
		}
		return nil
	}

	return writeChunk(file, data)
}

func writeChunk(file *os.File, data []byte) error {
	for len(data) > 0 {
		n, err := file.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return errors.New("printer accepted zero bytes")
		}
		data = data[n:]
	}
	return nil
}
