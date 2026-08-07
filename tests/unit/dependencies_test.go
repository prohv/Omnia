package unit

import (
	"runtime"
	"strings"
	"testing"

	"omnia/internal/dependencies"
)

func TestCheckDependencies(t *testing.T) {
	deps := dependencies.CheckDependencies()

	if len(deps) != 4 {
		t.Fatalf("expected 4 dependencies (3 native + 1 external), got %d", len(deps))
	}

	// Verify Native PDF Engine
	pdfDep := deps[0]
	if !pdfDep.IsNative || !pdfDep.Installed {
		t.Errorf("expected pdfcpu engine to be native and installed, got IsNative=%t, Installed=%t", pdfDep.IsNative, pdfDep.Installed)
	}

	// Verify Native Image Engine
	imgDep := deps[1]
	if !imgDep.IsNative || !imgDep.Installed {
		t.Errorf("expected image engine to be native and installed, got IsNative=%t, Installed=%t", imgDep.IsNative, imgDep.Installed)
	}

	// Verify Native OpenXML Engine
	openxmlDep := deps[2]
	if !openxmlDep.IsNative || !openxmlDep.Installed {
		t.Errorf("expected openxml engine to be native and installed, got IsNative=%t, Installed=%t", openxmlDep.IsNative, openxmlDep.Installed)
	}

	// Verify External Office Engine
	officeDep := deps[3]
	if officeDep.IsNative {
		t.Errorf("expected LibreOffice engine to be non-native")
	}

	if !officeDep.Installed {
		guide := officeDep.InstallGuide
		if guide == "" {
			t.Errorf("expected non-empty install guide for missing Office dependency")
		}
		switch runtime.GOOS {
		case "windows":
			if !strings.Contains(guide, "winget") {
				t.Errorf("expected windows install guide to mention winget, got %s", guide)
			}
		case "darwin":
			if !strings.Contains(guide, "brew") {
				t.Errorf("expected macOS install guide to mention brew, got %s", guide)
			}
		default:
			if !strings.Contains(guide, "apt") {
				t.Errorf("expected linux install guide to mention apt, got %s", guide)
			}
		}
	}
}
