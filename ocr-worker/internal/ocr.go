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

func runOCR(pdfPath string) (string, error) {

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

	final := TempName("ocr_output", ".txt")
	os.WriteFile(final, merged.Bytes(), 0644)

	log.Println("✅ OCR Completed:", final)
	return final, nil
}

// MAIN JOB PROCESSOR
func ProcessJob(job Job) {

	log.Println("⚙ OCR Worker processing:", job.Tool)
	UpdateStatus(job.ID, "processing")

	local := DownloadFromS3(job.Files[0])
	if local == "" {
			UpdateStatus(job.ID, "error")
			return
	}

	ext := strings.ToLower(filepath.Ext(local))

	var out string
	var err error

	switch job.Tool {

	case "ocr", "pdf-to-text":
			// PDF → TEXT
			out, err = runPDFOCR(local)

	case "image-to-text", "jpg-to-text", "png-to-text":
			// IMAGE → TEXT
			out, err = runImageOCR(local)

	default:
			log.Println("❌ Unknown OCR tool:", job.Tool)
			UpdateStatus(job.ID, "error")
			return
	}

	if err != nil || out == "" {
			log.Println("❌ OCR failed:", err)
			UpdateStatus(job.ID, "error")
			return
	}

	// Upload result
	url := UploadToS3(out)
	SaveResult(job.ID, url)

	DeleteFile(local)
	DeleteFile(out)

	UpdateStatus(job.ID, "completed")
	log.Println("✅ OCR job completed:", job.ID)
}
