package internal

import (
    "errors"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

// Auto-detect newest output file
func findNewestFile(ext string) (string, error) {
    pattern := filepath.Join("/tmp", "*"+ext)
    files, _ := filepath.Glob(pattern)

    if len(files) == 0 {
        return "", errors.New("no output files found for pattern " + pattern)
    }

    newest := files[0]
    newestTime := getMTime(newest)

    for _, f := range files[1:] {
        mt := getMTime(f)
        if mt.After(newestTime) {
            newest = f
            newestTime = mt
        }
    }

    return newest, nil
}

func getMTime(path string) time.Time {
    fi, err := os.Stat(path)
    if err != nil {
        return time.Time{}
    }
    return fi.ModTime()
}

// ----------------------------
// RUN LIBREOFFICE
// ----------------------------
func runLibreOffice(input, convertTo string) (string, error) {
    log.Println("🚀 Running LibreOffice:", convertTo, "→", input)

    cmd := exec.Command("soffice",
        "--headless",
        "--invisible",
        "--nodefault",
        "--nofirststartwizard",
        "--nologo",
        "--convert-to", convertTo,
        input,
        "--outdir", "/tmp",
    )

    if err := cmd.Run(); err != nil {
        log.Println("❌ LibreOffice failed:", err)
        return "", err
    }

    // Wait for output file to appear
    time.Sleep(800 * time.Millisecond)

    // Detect extension from convertTo
    ext := "." + convertTo
    if convertTo == "pdf" {
        ext = ".pdf"
    }
    if convertTo == "docx" {
        ext = ".docx"
    }
    if convertTo == "xlsx" {
        ext = ".xlsx"
    }
    if convertTo == "pptx" {
        ext = ".pptx"
    }

    // Auto-detect newest output file
    out, err := findNewestFile(ext)
    if err != nil {
        log.Println("❌ Output not detected:", err)
        return "", err
    }

    log.Println("✅ Detected output file:", out)
    return out, nil
}

// ----------------------------
// TOOL WRAPPERS
// ----------------------------
func officeToPDF(input string) (string, error) {
    return runLibreOffice(input, "pdf")
}

func pdfToWord(input string) (string, error) {
    return runLibreOffice(input, "docx")
}

func pdfToExcel(input string) (string, error) {
    return runLibreOffice(input, "xlsx")
}

func pdfToPPT(input string) (string, error) {
    return runLibreOffice(input, "pptx")
}

// ----------------------------
// MAIN JOB PROCESSOR
// ----------------------------
func ProcessJob(job Job) {

    log.Println("⚙ Processing Office job:", job.Tool)
    UpdateStatus(job.ID, "processing")

    input := DownloadFromS3(job.Files[0])
    if input == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    var output string
    var err error

    switch job.Tool {
    case "word-to-pdf", "excel-to-pdf", "ppt-to-pdf":
        output, err = officeToPDF(input)

    case "pdf-to-word":
        output, err = pdfToWord(input)

    case "pdf-to-excel":
        output, err = pdfToExcel(input)

    case "pdf-to-ppt":
        output, err = pdfToPPT(input)
    }

    if err != nil || output == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    // Upload
    finalURL := UploadToS3(output)
    if finalURL == "" {
        UpdateStatus(job.ID, "error")
        return
    }

    SaveResult(job.ID, finalURL)

    DeleteFile(input)
    DeleteFile(output)

    log.Println("✅ Office Job Completed:", job.ID)
}
