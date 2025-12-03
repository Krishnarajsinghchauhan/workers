package internal

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const magickPath = "/usr/bin/magick"

func findMagick() string {
	if _, err := exec.LookPath("magick"); err == nil {
		return "magick"
	}
	return magickPath
}

var MAGICK = findMagick()

// -----------------------------------
// IMAGE ENHANCER
// -----------------------------------
func enhanceScan(input string) string {
	out := TempName("enhanced", ".png")

	log.Println("🔧 Enhancing scan:", input)

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

	log.Println("✔ Enhanced scan →", out)
	return out
}

// -----------------------------------
// PDF → TEXT
// -----------------------------------
func runPDFOCR(pdfPath string) (string, error) {

	log.Println("📄 Step 1: Converting PDF → PNG pages...")

	base := "/tmp/ocr_page"

	cmd := exec.Command("pdftoppm", pdfPath, base, "-png", "-r", "300")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("❌ pdftoppm failed:", err)
		log.Println("Output:", string(out))
		return "", err
	}

	pages, _ := filepath.Glob(base + "-*.png")
	if len(pages) == 0 {
		return "", errors.New("no PNG pages produced")
	}

	sort.Strings(pages)

	var merged bytes.Buffer

	for _, pg := range pages {
		log.Println("🔍 OCR on:", pg)

		outBase := strings.TrimSuffix(pg, ".png")

		cmd := exec.Command("tesseract", pg, outBase, "--dpi", "300")
		tOut, tErr := cmd.CombinedOutput()

		if tErr != nil {
			log.Println("❌ Tesseract failed:", string(tOut))
			return "", tErr
		}

		txt, err := os.ReadFile(outBase + ".txt")
		if err == nil {
			merged.Write(txt)
			merged.WriteString("\n\n")
		}
	}

	final := TempFile("ocr_output", ".txt")
	os.WriteFile(final, merged.Bytes(), 0644)

	log.Println("✅ PDF OCR Completed:", final)
	return final, nil
}

// -----------------------------------
// IMAGE → TEXT
// -----------------------------------
func runImageOCR(imagePath string) (string, error) {

	log.Println("🖼  Running OCR on image:", imagePath)

	outBase := TempFile("image_ocr", "")

	cmd := exec.Command("tesseract", imagePath, outBase, "--dpi", "300")
	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("❌ tesseract image OCR failed:", string(data))
		return "", err
	}

	txtFile := outBase + ".txt"
	return txtFile, nil
}

// -----------------------------------
// MAIN JOB PROCESSOR
// -----------------------------------
func ProcessJob(job Job) {

	log.Println("⚙ OCR Worker processing:", job.Tool)
	UpdateStatus(job.ID, "processing")

	local := DownloadFromS3(job.Files[0])
	if local == "" {
		UpdateStatus(job.ID, "error")
		return
	}

	var out string
	var err error

	switch job.Tool {

	case "ocr", "pdf-to-text":
		out, err = runPDFOCR(local)

	case "image-to-text", "jpg-to-text", "png-to-text":
		out, err = runImageOCR(local)

	case "scanned-enhance":
		enhanced := enhanceScan(local)
		if enhanced == "" {
			UpdateStatus(job.ID, "error")
			return
		}
		out, err = runImageOCR(enhanced)
		DeleteFile(enhanced)

	default:
		log.Println("❌ Unknown OCR tool:", job.Tool)
		UpdateStatus(job.ID, "error")
		return
	}

	if err != nil || out == "" {
		UpdateStatus(job.ID, "error")
		return
	}

	url := UploadToS3(out)
	SaveResult(job.ID, url)

	DeleteFile(local)
	DeleteFile(out)

	UpdateStatus(job.ID, "completed")
	log.Println("✅ OCR job completed:", job.ID)
}
