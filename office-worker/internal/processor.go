package internal

import (
    "encoding/json"
    "errors"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

//
// ---------------------------------------------------
// FILE HELPERS
// ---------------------------------------------------
func findNewestFile(ext string) (string, error) {
    pattern := filepath.Join("/tmp", "*"+ext)
    files, _ := filepath.Glob(pattern)

    if len(files) == 0 {
        return "", errors.New("no output files found for " + pattern)
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

//
// ---------------------------------------------------
// LIBREOFFICE EXECUTION
// ---------------------------------------------------
func runLibreOffice(input, convertTo string) (string, error) {
    log.Println("🚀 LibreOffice converting:", input, "→", convertTo)

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

    time.Sleep(800 * time.Millisecond)

    // determine extension
    ext := "." + convertTo

    out, err := findNewestFile(ext)
    if err != nil {
        log.Println("❌ Output not found:", err)
        return "", err
    }

    log.Println("✅ LibreOffice output:", out)
    return out, nil
}

//
// ---------------------------------------------------
// PYTHON WORKER CALL
// ---------------------------------------------------
func RunPythonWorker(job Job) (string, error) {

    log.Println("🐍 Running Python worker for:", job.Tool)

    jsonBytes, _ := json.Marshal(map[string]interface{}{
        "job_id": job.ID,
        "tool":   job.Tool,
        "files":  job.Files,
    })

    cmd := exec.Command("python3", "/home/ubuntu/code/workers/office-python-worker/worker.py")
    stdin, _ := cmd.StdinPipe()

    stdin.Write(jsonBytes)
    stdin.Close()

    out, err := cmd.CombinedOutput()
    if err != nil {
        log.Println("❌ Python worker failed:", err, "Output:", string(out))
        return "", err
    }

    var response map[string]string
    json.Unmarshal(out, &response)

    // Python returns final S3 URL
    return response["url"], nil
}

//
// ---------------------------------------------------
// MAIN JOB PROCESSOR
// ---------------------------------------------------
func ProcessJob(job Job) {

    log.Println("⚙ Processing job:", job.Tool)
    UpdateStatus(job.ID, "processing")

    var input string
    var output string
    var err error

    // -------------------------
    // 1) LibreOffice jobs
    // -------------------------
    if job.Tool == "word-to-pdf" || job.Tool == "excel-to-pdf" || job.Tool == "ppt-to-pdf" {

        input = DownloadFromS3(job.Files[0])
        if input == "" {
            UpdateStatus(job.ID, "error")
            return
        }

        output, err = runLibreOffice(input, "pdf")
        if err != nil {
            UpdateStatus(job.ID, "error")
            return
        }

        finalURL := UploadToS3(output)
        if finalURL == "" {
            UpdateStatus(job.ID, "error")
            return
        }

        SaveResult(job.ID, finalURL)

        DeleteFile(input)
        DeleteFile(output)

        log.Println("✅ LibreOffice Job Completed:", job.ID)
        return
    }

    // -------------------------
    // 2) Python worker jobs
    // -------------------------
    if job.Tool == "pdf-to-word" || job.Tool == "pdf-to-excel" || job.Tool == "pdf-to-ppt" {

        // Python worker handles S3 download + upload
        output, err = RunPythonWorker(job)
        if err != nil || output == "" {
            UpdateStatus(job.ID, "error")
            return
        }

        // Output IS ALREADY an S3 URL
        SaveResult(job.ID, output)

        log.Println("✅ Python Job Completed:", job.ID)
        return
    }

    // -------------------------
    // INVALID TOOL
    // -------------------------
    log.Println("❌ Unknown tool:", job.Tool)
    UpdateStatus(job.ID, "error")
}
