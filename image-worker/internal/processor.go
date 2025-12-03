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

    log.Println("📄 Converting PDF to images...")

    outPrefix := "/tmp/ocr_page"

    // Convert PDF → PNG (multi-page)
    cmd := exec.Command("pdftoppm", pdfPath, outPrefix, "-png")
    if out, err := cmd.CombinedOutput(); err != nil {
        log.Println("❌ pdftoppm failed:", err)
        log.Println("Output:", string(out))
        return "", err
    }

    // Find all PNG pages
    pages, _ := filepath.Glob(outPrefix + "-*.png")
    if len(pages) == 0 {
        return "", errors.New("no PNG pages created from PDF")
    }

    sort.Strings(pages)

    // OCR all pages
    var buf bytes.Buffer

    for _, img := range pages {
        log.Println("🔍 OCR:", img)

        base := strings.TrimSuffix(img, ".png")

        cmd := exec.Command("tesseract", img, base, "--dpi", "300")
        if out, err := cmd.CombinedOutput(); err != nil {
            log.Println("❌ tesseract failed:", string(out))
            return "", err
        }

        txtPath := base + ".txt"
        data, _ := os.ReadFile(txtPath)

        buf.Write(data)
        buf.WriteString("\n\n")
    }

    // Save final merged text file
    final := TempFile("ocr", ".txt")
    os.WriteFile(final, buf.Bytes(), 0644)

    log.Println("✅ OCR output ready:", final)
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
