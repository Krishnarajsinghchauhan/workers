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

	// Convert PDF to PNG images (one per page)
	cmd := exec.Command("pdftoppm", pdfPath, base, "-png")
	out, err := cmd.CombinedOutput()
	if err != nil {
			log.Println("❌ pdftoppm failed:", err)
			log.Println("Output:", string(out))
			return "", err
	}

	// Collect generated PNGs
	pages, _ := filepath.Glob(base + "-*.png")
	if len(pages) == 0 {
			log.Println("❌ No pages produced by pdftoppm")
			return "", errors.New("no PNG pages created")
	}

	sort.Strings(pages)

	log.Println("📄 Pages generated:", pages)

	var buf bytes.Buffer

	// Run OCR on each PNG page
	for _, pg := range pages {
			log.Println("🔍 Running Tesseract on:", pg)

			outTxtBase := strings.TrimSuffix(pg, ".png")
			cmd := exec.Command("tesseract", pg, outTxtBase, "--dpi", "300")

			tOut, tErr := cmd.CombinedOutput()
			if tErr != nil {
					log.Println("❌ Tesseract failed:", string(tOut))
					return "", tErr
			}

			txtData, err := os.ReadFile(outTxtBase + ".txt")
			if err == nil {
					buf.Write(txtData)
					buf.WriteString("\n\n")
			}
	}

	// Save final merged OCR file
	final := TempFile("ocr_output", ".txt")
	os.WriteFile(final, buf.Bytes(), 0644)

	log.Println("✅ OCR Completed:", final)
	return final, nil
}


func ProcessJob(job Job) {

    log.Println("⚙ OCR Worker processing:", job.Tool)
    UpdateStatus(job.ID, "processing")

    pdfFile := DownloadFromS3(job.Files[0])
    if pdfFile == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    out, err := runOCR(pdfFile)
    if err != nil {
        log.Println("❌ runOCR failed:", err)
        UpdateStatus(job.ID, "error")
        return
    }

    url := UploadToS3(out)
    if url == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    SaveResult(job.ID, url)

    DeleteFile(pdfFile)
    DeleteFile(out)

    log.Println("✅ OCR Job Completed:", job.ID)
}
