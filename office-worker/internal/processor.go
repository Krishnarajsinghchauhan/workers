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

func RunPythonWorker(job Job) (string, error) {

    log.Println("🐍 Python Worker triggered for:", job.Tool)

    jsonBytes, _ := json.Marshal(map[string]interface{}{
        "job_id": job.ID,
        "tool":   job.Tool,
        "files":  job.Files,
    })

    cmd := exec.Command("python3", "/home/ubuntu/office-python-worker/worker.py")
    stdin, _ := cmd.StdinPipe()

    stdin.Write(jsonBytes)
    stdin.Close()

    out, err := cmd.CombinedOutput()
    if err != nil {
        log.Println("❌ Python worker failed:", err, string(out))
        return "", err
    }

    var response map[string]string
    json.Unmarshal(out, &response)

    return response["url"], nil
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

    case "pdf-to-word", "pdf-to-excel", "pdf-to-ppt":
        output, err = RunPythonWorker(job)
    

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
