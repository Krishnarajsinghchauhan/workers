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

    log.Println("📄 Step 1: Converting PDF → PNG pages...")

    base := "/tmp/ocr_page"

    // Convert PDF to PNG pages
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
    log.Println("📄 PNG pages:", pages)

    var merged bytes.Buffer

    for _, pg := range pages {
        log.Println("🔍 OCR on:", pg)

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
            return "", tErr
        }

        txt, err := os.ReadFile(outBase + ".txt")
        if err == nil {
            merged.Write(txt)
            merged.WriteString("\n\n")
        }
    }

    // Final output file
    final := TempFile("ocr_output", ".txt")
    os.WriteFile(final, merged.Bytes(), 0644)

    log.Println("✅ OCR completed:", final)
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
        UpdateStatus(job.ID, "error")
        log.Println("❌ runOCR failed:", err)
        return
    }

    finalURL := UploadToS3(output)
    SaveResult(job.ID, finalURL)
    UpdateStatus(job.ID, "completed")

    DeleteFile(local)
    DeleteFile(output)

    log.Println("✅ OCR Job Completed:", job.ID)
}
