package internal

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

const magickPath = "/opt/homebrew/bin/magick"

func findMagick() string {
	if _, err := exec.LookPath("magick"); err == nil {
		return "magick"
	}
	return magickPath
}

var MAGICK = findMagick()

// CLEAN IMAGE
func preprocess(input string) string {
	out := TempName("clean", ".png")

	log.Println("➡ Running preprocess using:", MAGICK)

	cmd := exec.Command(
		MAGICK,
		input,
		"-alpha", "remove",
		"-strip",
		out,
	)

	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println("⚠️ preprocess failed:", err)
		return input
	}

	log.Println("✔ preprocess OK →", out)
	return out
}

// IMAGE → TEXT
func extractText(input string) string {
	clean := preprocess(input)
	out := TempName("ocr_text", "")

	cmd := exec.Command(
		"tesseract",
		clean,
		out,
		"--psm", "3",
	)

	raw, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("❌ extractText failed:", err)
		log.Println("🔍 Tesseract Output:", string(raw))
		return ""
	}

	return out + ".txt"
}

// IMAGE/PDF → SEARCHABLE PDF
func runOCR(input string) string {
	out := TempName("ocr_pdf", ".pdf")

	cmd := exec.Command(
		"tesseract",
		input,
		strings.TrimSuffix(out, ".pdf"),
		"pdf",
	)

	raw, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("❌ runOCR failed:", err)
		log.Println("🔍 Tesseract Output:", string(raw))
		return ""
	}

	return out
}

// ENHANCE SCAN
func enhanceScan(input string) string {
	out := TempName("enhanced", ".png")

	cmd := exec.Command(
		MAGICK,
		input,
		"-normalize",
		"-brightness-contrast", "10x20",
		out,
	)

	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println("❌ enhanceScan failed:", err)
		return ""
	}

	return out
}
