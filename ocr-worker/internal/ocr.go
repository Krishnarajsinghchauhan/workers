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

func detectImageMagick() string {
	// Linux standard: convert
	if path, err := exec.LookPath("convert"); err == nil {
			log.Println("🪄 Using ImageMagick:", path)
			return path
	}

	// macOS Homebrew
	if _, err := os.Stat("/opt/homebrew/bin/magick"); err == nil {
			log.Println("🪄 Using Homebrew magick")
			return "/opt/homebrew/bin/magick"
	}

	log.Println("❌ No ImageMagick found!")
	return ""
}

var MAGICK = detectImageMagick()


func findMagick() string {
	if _, err := exec.LookPath("magick"); err == nil {
		return "magick"
	}
	return magickPath
}



// -----------------------------------
// IMAGE ENHANCER
// -----------------------------------
func enhancePDF(pdfPath string) (string, error) {

	log.Println("📄 Step 1: PDF → PNG pages (300 DPI)")

	base := "/tmp/enh_page"
	cmd := exec.Command("pdftoppm", pdfPath, base, "-png", "-r", "300")
	out, err := cmd.CombinedOutput()
	if err != nil {
			log.Println("❌ pdftoppm failed:", string(out))
			return "", err
	}

	pages, _ := filepath.Glob(base + "-*.png")
	if len(pages) == 0 {
			return "", errors.New("no PNG pages extracted")
	}

	sort.Strings(pages)

	log.Println("📄 Found pages:", pages)

	enhancedPages := []string{}

	for _, pg := range pages {
			outPg := strings.TrimSuffix(pg, ".png") + "_enh.png"

			log.Println("🔧 Enhancing:", pg)

			cmd := exec.Command(
					MAGICK,
					pg,
					"-normalize",
					"-brightness-contrast", "10x20",
					outPg,
			)

			if err := cmd.Run(); err != nil {
					log.Println("❌ Enhance failed:", err)
					return "", err
			}

			enhancedPages = append(enhancedPages, outPg)
	}

	// Output PDF
	finalPDF := TempFile("enhanced_pdf", ".pdf")

	log.Println("📄 Step 3: combining pages → PDF:", finalPDF)

	args := append(enhancedPages, finalPDF)
	cmd = exec.Command("convert", args...)
	if err := cmd.Run(); err != nil {
			log.Println("❌ convert to PDF failed:", err)
			return "", err
	}

	log.Println("✅ Enhanced PDF created:", finalPDF)
	return finalPDF, nil
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
