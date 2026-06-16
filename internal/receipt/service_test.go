package receipt

import (
	"strings"
	"testing"
)

func TestServiceBuildFormatsReceipt(t *testing.T) {
	service := NewService(ModeESCPOS)

	got, err := service.Build(Request{
		StoreName: "TOKO PUTRI",
		Items: []Item{
			{Name: "Nasi Goreng", Qty: 1, Price: 15000},
		},
		Total: 15000,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	receipt := string(got)
	assertContains(t, receipt, cmdInit+cmdCenter+cmdBoldOn+"TOKO PUTRI"+newLine+cmdBoldOff)
	assertContains(t, receipt, cmdLeft+"------------------------"+newLine)
	assertContains(t, receipt, "Nasi Goreng       15.000"+newLine)
	assertContains(t, receipt, "TOTAL             15.000"+newLine)
	assertContains(t, receipt, "TERIMA KASIH"+newLine)

	if !strings.HasSuffix(receipt, cmdCut) {
		t.Fatalf("receipt should end with cut command %q, got %q", cmdCut, receipt[len(receipt)-len(cmdCut):])
	}
}

func TestServiceBuildValidatesInput(t *testing.T) {
	service := NewService(ModeESCPOS)

	_, err := service.Build(Request{})
	if err == nil {
		t.Fatal("Build() expected validation error")
	}
}

func TestServiceBuildFormatsPlainReceipt(t *testing.T) {
	service := NewService(ModePlain)

	got, err := service.Build(Request{
		StoreName: "TOKO PUTRI",
		Items: []Item{
			{Name: "Nasi Goreng", Qty: 1, Price: 15000},
		},
		Total: 15000,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	receipt := string(got)
	assertContains(t, receipt, "       TOKO PUTRI\n")
	assertContains(t, receipt, "------------------------\n")
	assertContains(t, receipt, "Nasi Goreng       15.000\n")
	assertContains(t, receipt, "TOTAL             15.000\n")
	assertContains(t, receipt, "      TERIMA KASIH\n")
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("receipt missing %q in %q", needle, haystack)
	}
}
