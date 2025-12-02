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
func reorderPDF(input string, opts map[string]string) string {
	order := opts["order"]
	if order == "" {
			log.Println("❌ reorderPDF: no page order provided")
			return ""
	}

	// Clean spaces
	order = strings.ReplaceAll(order, " ", "")

	// Split page numbers: "3,1,2" → ["3","1","2"]
	pageOrder := strings.Split(order, ",")

	// Convert to ABSOLUTE PATHS (important)
	absIn, err := filepath.Abs(input)
	if err != nil {
			log.Println("❌ Failed to get abs input path:", err)
			return ""
	}

	// Step 1 — Repair PDF
	repaired := TempName("fixed", ".pdf")
	absRepaired, _ := filepath.Abs(repaired)

	repairCmd := exec.Command("qpdf", "--linearize", absIn, absRepaired)
	repairOutput, repairErr := repairCmd.CombinedOutput()

	if repairErr != nil {
			if exitErr, ok := repairErr.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
					log.Println("⚠️ qpdf repair warnings — continuing")
			} else {
					log.Println("❌ qpdf repair failed:", repairErr)
					log.Println("Output:", string(repairOutput))
					return ""
			}
	}

	// Step 2 — Build reorder command (correct qpdf syntax)
	outFile := TempName("reordered", ".pdf")
	absOut, _ := filepath.Abs(outFile)

	// Final qpdf args:
	// qpdf repaired.pdf output.pdf --pages repaired.pdf 3 1 2 --
	args := []string{
			absRepaired,  // input
			absOut,       // output
			"--pages",
			absRepaired,  // input again
	}

	// Add page numbers individually (CRITICAL)
	// Example: ["3","1","2"]
	args = append(args, pageOrder...)

	// End token
	args = append(args, "--")

	log.Println("➡ qpdf reorder args:", args)

	// Execute the command
	reorderCmd := exec.Command("qpdf", args...)
	output, err := reorderCmd.CombinedOutput()

	if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
					log.Println("⚠️ qpdf reorder warnings — continuing")
			} else {
					log.Println("❌ qpdf reorder failed:", err)
					log.Println("Output:", string(output))
					return ""
			}
	}

	log.Println("✅ PDF reordered:", absOut)
	return absOut
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
