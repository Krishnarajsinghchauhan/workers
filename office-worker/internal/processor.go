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
// FILE HELPERS
//
func findNewestFile(ext string) (string, error) {
    pattern := filepath.Join("/tmp", "*"+ext)
    files, _ := filepath.Glob(pattern)
    if len(files) == 0 {
        return "", errors.New("no output files found: " + pattern)
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
// LIBREOFFICE
//
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

    out, err := findNewestFile("." + convertTo)
    if err != nil {
        log.Println("❌ Output not found:", err)
        return "", err
    }

    log.Println("✅ LibreOffice output:", out)
    return out, nil
}

//
// PYTHON WORKER
//
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

    out, _ := cmd.CombinedOutput() // <-- DO NOT check err here!
    log.Println("🐍 Python Output:\n" + string(out))

    var response map[string]string
    if err := json.Unmarshal(out, &response); err != nil {
        return "", errors.New("Invalid JSON from Python: " + string(out))
    }

    if status, ok := response["status"]; ok && status == "completed" {
        return response["url"], nil
    }

    if response["error"] != "" {
        return "", errors.New(response["error"])
    }

    return "", errors.New("Unknown python output")
}



//
// PROCESS JOB
//
func ProcessJob(job Job) {

    log.Println("⚙ Processing job:", job.Tool)
    UpdateStatus(job.ID, "processing")

    //
    // LIBREOFFICE JOBS
    //
    if job.Tool == "word-to-pdf" || job.Tool == "excel-to-pdf" || job.Tool == "ppt-to-pdf" {

        input := DownloadFromS3(job.Files[0])
        if input == "" {
            UpdateStatus(job.ID, "error")
            return
        }

        output, err := runLibreOffice(input, "pdf")
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

    //
    // PYTHON JOBS
    //
    if job.Tool == "pdf-to-word" || job.Tool == "pdf-to-excel" || job.Tool == "pdf-to-ppt" {

        url, err := RunPythonWorker(job)
        if err != nil {
            log.Println("❌ Python worker failed:", err)
            UpdateStatus(job.ID, "error")
            return
        }

        SaveResult(job.ID, url)

        log.Println("✅ Python Job Completed:", job.ID)
        return
    }

    log.Println("❌ Unknown tool:", job.Tool)
    UpdateStatus(job.ID, "error")
}
