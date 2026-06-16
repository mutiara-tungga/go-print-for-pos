package receipt

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	cmdInit    = "\x1B\x40"
	cmdCenter  = "\x1B\x61\x01"
	cmdLeft    = "\x1B\x61\x00"
	cmdBoldOn  = "\x1B\x45\x01"
	cmdBoldOff = "\x1B\x45\x00"
	cmdCut     = "\x1D\x56\x00"

	newLine      = "\r\n"
	receiptWidth = 24
)

type Request struct {
	StoreName string `json:"store_name"`
	Items     []Item `json:"items"`
	Total     int64  `json:"total"`
}

type Item struct {
	Name  string `json:"name"`
	Qty   int64  `json:"qty"`
	Price int64  `json:"price"`
}

type Mode string

const (
	ModeESCPOS Mode = "escpos"
	ModePlain  Mode = "plain"
)

type Service struct {
	mode Mode
}

func NewService(mode Mode) *Service {
	if mode == "" {
		mode = ModeESCPOS
	}
	return &Service{mode: mode}
}

func (s *Service) Build(req Request) ([]byte, error) {
	if err := validate(req); err != nil {
		return nil, err
	}

	if s.mode == ModePlain {
		return s.buildPlain(req), nil
	}

	return s.buildESCPOS(req), nil
}

func (s *Service) buildESCPOS(req Request) []byte {
	var buf bytes.Buffer
	separator := strings.Repeat("-", receiptWidth)

	buf.WriteString(cmdInit)
	buf.WriteString(cmdCenter)
	buf.WriteString(cmdBoldOn)
	writeLine(&buf, req.StoreName)
	buf.WriteString(cmdBoldOff)
	writeLine(&buf, "")

	buf.WriteString(cmdLeft)
	writeLine(&buf, separator)
	for _, item := range req.Items {
		name := item.Name
		if item.Qty > 1 {
			name = fmt.Sprintf("%dx %s", item.Qty, item.Name)
		}
		writeLine(&buf, alignLine(name, formatRupiah(item.Qty*item.Price), receiptWidth))
	}
	writeLine(&buf, separator)
	buf.WriteString(cmdBoldOn)
	writeLine(&buf, alignLine("TOTAL", formatRupiah(req.Total), receiptWidth))
	buf.WriteString(cmdBoldOff)

	writeLine(&buf, "")
	buf.WriteString(cmdCenter)
	writeLine(&buf, "TERIMA KASIH")
	writeLine(&buf, "")
	writeLine(&buf, "")
	writeLine(&buf, "")
	buf.WriteString(cmdCut)

	return buf.Bytes()
}

func (s *Service) buildPlain(req Request) []byte {
	var buf bytes.Buffer
	separator := strings.Repeat("-", receiptWidth)

	writePlainLine(&buf, centerText(req.StoreName, receiptWidth))
	writePlainLine(&buf, "")
	writePlainLine(&buf, separator)
	for _, item := range req.Items {
		name := item.Name
		if item.Qty > 1 {
			name = fmt.Sprintf("%dx %s", item.Qty, item.Name)
		}
		writePlainLine(&buf, alignLine(name, formatRupiah(item.Qty*item.Price), receiptWidth))
	}
	writePlainLine(&buf, separator)
	writePlainLine(&buf, alignLine("TOTAL", formatRupiah(req.Total), receiptWidth))
	writePlainLine(&buf, "")
	writePlainLine(&buf, centerText("TERIMA KASIH", receiptWidth))
	writePlainLine(&buf, "")
	writePlainLine(&buf, "")

	return buf.Bytes()
}

func validate(req Request) error {
	if strings.TrimSpace(req.StoreName) == "" {
		return errors.New("store_name is required")
	}
	if len(req.Items) == 0 {
		return errors.New("items is required")
	}
	if req.Total < 0 {
		return errors.New("total must be greater than or equal to 0")
	}
	for i, item := range req.Items {
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("items[%d].name is required", i)
		}
		if item.Qty <= 0 {
			return fmt.Errorf("items[%d].qty must be greater than 0", i)
		}
		if item.Price < 0 {
			return fmt.Errorf("items[%d].price must be greater than or equal to 0", i)
		}
	}
	return nil
}

func writeLine(buf *bytes.Buffer, value string) {
	buf.WriteString(value)
	buf.WriteString(newLine)
}

func writePlainLine(buf *bytes.Buffer, value string) {
	buf.WriteString(value)
	buf.WriteByte('\n')
}

func centerText(value string, width int) string {
	value = strings.TrimSpace(value)
	if len(value) >= width {
		return value
	}
	padding := (width - len(value)) / 2
	return strings.Repeat(" ", padding) + value
}

func alignLine(left, right string, width int) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)

	if len(left)+len(right)+1 > width {
		maxLeft := width - len(right) - 1
		if maxLeft < 1 {
			return right
		}
		left = truncate(left, maxLeft)
	}

	spaces := width - len(left) - len(right)
	if spaces < 1 {
		spaces = 1
	}
	return left + strings.Repeat(" ", spaces) + right
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func formatRupiah(value int64) string {
	raw := strconv.FormatInt(value, 10)
	if len(raw) <= 3 {
		return raw
	}

	var parts []string
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	parts = append([]string{raw}, parts...)
	return strings.Join(parts, ".")
}
