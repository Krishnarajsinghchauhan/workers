package internal

import (
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// Build correct LibreOffice output path
func outputPath(input, newExt string) string {
    base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
    return filepath.Join("/tmp", base+newExt)
}

// ----------------------------
// Convert Office → PDF
// ----------------------------
func officeToPDF(input string) string {
    output := outputPath(input, ".pdf")

    cmd := exec.Command("soffice",
        "--headless",
        "--convert-to", "pdf",
        input,
        "--outdir", "/tmp",
    )

    if err := cmd.Run(); err != nil {
        log.Println("❌ Office-to-PDF failed:", err)
        return ""
    }

    return output
}

// ----------------------------
// PDF → Word
// ----------------------------
func pdfToWord(input string) string {
    output := outputPath(input, ".docx")

    cmd := exec.Command("soffice",
        "--headless",
        "--convert-to", "docx",
        input,
        "--outdir", "/tmp",
    )

    if err := cmd.Run(); err != nil {
        log.Println("❌ PDF-to-Word failed:", err)
        return ""
    }

    return output
}

// ----------------------------
// PDF → Excel
// ----------------------------
func pdfToExcel(input string) string {
    output := outputPath(input, ".xlsx")

    cmd := exec.Command("soffice",
        "--headless",
        "--convert-to", "xlsx",
        input,
        "--outdir", "/tmp",
    )

    if err := cmd.Run(); err != nil {
        log.Println("❌ PDF-to-Excel failed:", err)
        return ""
    }

    return output
}

// ----------------------------
// PDF → PowerPoint
// ----------------------------
func pdfToPPT(input string) string {
    output := outputPath(input, ".pptx")

    cmd := exec.Command("soffice",
        "--headless",
        "--convert-to", "pptx",
        input,
        "--outdir", "/tmp",
    )

    if err := cmd.Run(); err != nil {
        log.Println("❌ PDF-to-PPT failed:", err)
        return ""
    }

    return output
}

// ----------------------------
// MAIN JOB PROCESSOR
// ----------------------------
func ProcessJob(job Job) {

    log.Println("⚙ Processing Office job:", job.Tool)

    UpdateStatus(job.ID, "processing")

    // 1. Download input file
    input := DownloadFromS3(job.Files[0])
    if input == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    var output string

    switch job.Tool {
    case "word-to-pdf", "excel-to-pdf", "ppt-to-pdf":
        output = officeToPDF(input)

    case "pdf-to-word":
        output = pdfToWord(input)

    case "pdf-to-excel":
        output = pdfToExcel(input)

    case "pdf-to-ppt":
        output = pdfToPPT(input)
    }

    if output == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    // MUST check that file exists
    if _, err := os.Stat(output); err != nil {
        log.Println("❌ Converted file NOT FOUND:", output)
        UpdateStatus(job.ID, "error")
        return
    }

    // 3. Upload to S3
    finalURL := UploadToS3(output)
    if finalURL == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    SaveResult(job.ID, finalURL)

    // Cleanup
    DeleteFile(input)
    DeleteFile(output)

    log.Println("✅ Job completed:", job.ID)
}
