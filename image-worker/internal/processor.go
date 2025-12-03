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

// --------------
// OCR PROCESSING
// --------------
func runOCR(pdfPath string) (string, error) {

    log.Println("📄 Step 1: Converting PDF → PNG pages using pdftoppm...")

    base := "/tmp/ocr_page"

    // Convert PDF to PNG pages
    cmd := exec.Command("pdftoppm", pdfPath, base, "-png", "-r", "300")
    out, err := cmd.CombinedOutput()
    if err != nil {
        log.Println("❌ pdftoppm failed:", err)
        log.Println("Output:", string(out))
        return "", errors.New("pdftoppm failed")
    }

    // Get the page list
    pages, _ := filepath.Glob(base + "-*.png")
    if len(pages) == 0 {
        log.Println("❌ No PNG pages generated!")
        return "", errors.New("no PNG pages produced")
    }

    sort.Strings(pages)
    log.Println("📄 PNG pages generated:", pages)

    var merged bytes.Buffer

    // OCR page-by-page
    for _, pg := range pages {
        log.Println("🔍 Running OCR on:", pg)

        outBase := strings.TrimSuffix(pg, ".png")

        cmd := exec.Command(
            "tesseract",
            pg,
            outBase,
            "--dpi", "300",
        )

        tOut, tErr := cmd.CombinedOutput()
        if tErr != nil {
            log.Println("❌ Tesseract failed:", string(tOut))
            return "", errors.New("tesseract failed on page " + pg)
        }

        txtFile := outBase + ".txt"
        text, err := os.ReadFile(txtFile)
        if err != nil {
            log.Println("⚠️ Could not read:", txtFile)
            continue
        }

        merged.Write(text)
        merged.WriteString("\n\n")
    }

    // Save final merged txt
    final := TempFile("ocr_output", ".txt")
    os.WriteFile(final, merged.Bytes(), 0644)

    log.Println("✅ OCR completed. Output file:", final)
    return final, nil
}

// --------------
// MAIN PROCESSOR
// --------------
func ProcessJob(job Job) {
    log.Println("⚙ OCR Worker processing:", job.Tool)

    UpdateStatus(job.ID, "processing")

    local := DownloadFromS3(job.Files[0])
    if local == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    output, err := runOCR(local)
    if err != nil {
        log.Println("❌ runOCR failed:", err)
        UpdateStatus(job.ID, "error")
        return
    }

    url := UploadToS3(output)
    SaveResult(job.ID, url)

    UpdateStatus(job.ID, "completed")

    DeleteFile(local)
    DeleteFile(output)

    log.Println("✅ OCR Job Completed:", job.ID)
}
