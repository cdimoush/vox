package clipboard

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	tool, err := Detect()
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if tool == "" {
		t.Fatal("Detect() returned empty string")
	}
}

func TestDetectReturnsPbcopyOnDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping: not running on darwin")
	}
	tool, err := Detect()
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if tool != "pbcopy" {
		t.Errorf("expected pbcopy on darwin, got %q", tool)
	}
}

func TestWrite(t *testing.T) {
	err := Write("vox test string")
	if err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}
}
