package internal

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// ----------------------------
// MERGE PDFs
// ----------------------------
func mergePDFs(files []string) string {
	out := TempName("merged", ".pdf")
	args := append(files, out)

	if exec.Command("pdfunite", args...).Run() == nil {
		return out
	}

	log.Println("❌ pdfunite missing or failed, cannot merge")
	return ""
}

// ----------------------------
// SPLIT PDF
// ----------------------------
func splitPDF(input string, opts map[string]string) []string {
	outPattern := TempName("split", "-%02d.pdf")

	if exec.Command("pdfseparate", input, outPattern).Run() != nil {
		log.Println("❌ pdfseparate missing or failed")
		return nil
	}

	pattern := strings.Replace(outPattern, "%02d", "*", 1)
	files, _ := filepath.Glob(pattern)
	return files
}

// ----------------------------
// COMPRESS PDF
// ----------------------------
func compressPDF(input string) string {
	out := TempName("compressed", ".pdf")

	cmd := exec.Command("gs",
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=/ebook",
		"-dNOPAUSE", "-dQUIET", "-dBATCH",
		"-sOutputFile="+out,
		input,
	)

	if cmd.Run() != nil {
		log.Println("❌ Ghostscript missing or failed")
		return ""
	}

	return out
}

// ----------------------------
// ROTATE PDF
// ----------------------------
func rotatePDF(input string, opts map[string]string) string {
	// ----------------------------
	// 1. Validate rotation angle
	// ----------------------------
	angle := opts["angle"]
	if angle == "" {
			angle = "90"
	}

	// ----------------------------
	// 2. Convert to absolute paths (CRITICAL on Linux workers)
	// ----------------------------
	absIn, err := filepath.Abs(input)
	if err != nil {
			log.Println("❌ Failed to get absolute input path:", err)
			return ""
	}

	out := TempName("rotated", ".pdf")
	absOut, err := filepath.Abs(out)
	if err != nil {
			log.Println("❌ Failed to get absolute output path:", err)
			return ""
	}

	// ----------------------------
	// 3. Validate the PDF before rotating
	// ----------------------------
	check := exec.Command("qpdf", "--check", absIn)
	if err := check.Run(); err != nil {
			log.Println("❌ Invalid PDF input for rotation:", absIn, "Error:", err)
			return ""
	}

	// ----------------------------
	// 4. Perform rotation
	// ----------------------------
	cmd := exec.Command("qpdf", "--rotate="+angle, absIn, absOut)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
			log.Println("❌ qpdf rotation failed")
			log.Println("Command:", cmd.String())
			log.Println("Output:", string(outBytes))
			return ""
	}

	// ----------------------------
	// 5. Success
	// ----------------------------
	log.Println("✅ PDF rotated:", absOut)
	return absOut
}

// ----------------------------
// REORDER PDF (change page order)
// ----------------------------
// opts["order"] should be "3,1,2" for example
func reorderPDF(input string, opts map[string]string) string {
	order := opts["order"]
	if order == "" {
			log.Println("❌ reorderPDF: no page order provided")
			return ""
	}

	// Step 1 — Repair PDF
	repaired := TempName("fixed", ".pdf")
	repairCmd := exec.Command("qpdf", "--linearize", input, repaired)

	repairOutput, repairErr := repairCmd.CombinedOutput()

	if repairErr != nil {
			if exitErr, ok := repairErr.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
					// Exit code 3 = success with warnings → allowed
					log.Println("⚠️ qpdf repair succeeded with warnings — continuing")
			} else {
					log.Println("❌ qpdf repair failed:", repairErr)
					log.Println("Output:", string(repairOutput))
					return ""
			}
	}

	// Step 2 — Reorder PDF
	// Step 2 — Reorder PDF
outFile := TempName("reordered", ".pdf")
pageOrder := strings.Split(order, ",")

// qpdf syntax:
// qpdf input.pdf output.pdf --pages input.pdf 1 3 2 --
args := []string{
    repaired,         // input
    outFile,          // output
    "--pages",
    repaired,         // input again after --pages
}

// append page numbers
args = append(args, pageOrder...)

// add final separator
args = append(args, "--")

log.Println("➡ qpdf reorder args:", args)

reorderCmd := exec.Command("qpdf", args...)
reorderOutput, reorderErr := reorderCmd.CombinedOutput()

if reorderErr != nil {
    if exitErr, ok := reorderErr.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
        log.Println("⚠️ reorder succeeded with warnings — continuing")
    } else {
        log.Println("❌ qpdf reorder failed:", reorderErr)
        log.Println("Output:", string(reorderOutput))
        return ""
    }
}

	log.Println("✅ PDF reordered:", outFile)
	return outFile
}






// ----------------------------
// MAIN PROCESSOR
// ----------------------------
func ProcessJob(job Job) {
	log.Println("⚙ Processing PDF job:", job.Tool)
	UpdateStatus(job.ID, "processing")

	// 1. Download all input PDFs
	var local []string
	for _, f := range job.Files {
		p := DownloadFromS3(f)
		if p == "" {
			UpdateStatus(job.ID, "error")
			return
		}
		local = append(local, p)
	}

	var outputs []string

	// 2. Process
	switch job.Tool {

	case "merge":
		out := mergePDFs(local)
		if out != "" {
			outputs = []string{out}
		}

	case "split":
		outputs = splitPDF(local[0], job.Options)

	case "compress":
		out := compressPDF(local[0])
		if out != "" {
			outputs = []string{out}
		}

	case "rotate":
		out := rotatePDF(local[0], job.Options)
		if out != "" {
			outputs = []string{out}
		}

	case "reorder":
    out := reorderPDF(local[0], job.Options)
    if out != "" {
        outputs = []string{out}
    }

	default:
		log.Println("❌ Unknown tool:", job.Tool)
		UpdateStatus(job.ID, "error")
		return
	}

	// 3. Validate results
	if len(outputs) == 0 {
		UpdateStatus(job.ID, "error")
		log.Println("❌ No output generated")
		return
	}

	// 4. Upload to S3
	var urls []string
	for _, out := range outputs {
		url := UploadToS3(out)
		if url == "" {
			UpdateStatus(job.ID, "error")
			return
		}
		urls = append(urls, url)
		DeleteFile(out)
	}

	// 5. Cleanup input files
	for _, p := range local {
		DeleteFile(p)
	}

	// 6. Save result
	SaveResult(job.ID, urls)


	UpdateStatus(job.ID, "completed")
	log.Println("✅ PDF job completed:", job.ID)
}
