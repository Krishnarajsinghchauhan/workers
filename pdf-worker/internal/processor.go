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

	order = strings.ReplaceAll(order, " ", "")

	absIn, err := filepath.Abs(input)
	if err != nil {
			log.Println("❌ abs path error:", err)
			return ""
	}

	// Repair first
	repaired := TempName("fixed", ".pdf")
	absRepaired, _ := filepath.Abs(repaired)

	repair := exec.Command("qpdf", "--linearize", absIn, absRepaired)
	out, err := repair.CombinedOutput()

	if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
					log.Println("⚠️ qpdf repair warnings — continuing")
			} else {
					log.Println("❌ qpdf repair failed:", err)
					log.Println("Output:", string(out))
					return ""
			}
	}

	// FINAL output file
	outFile := TempName("reordered", ".pdf")
	absOut, _ := filepath.Abs(outFile)

	// UNIVERSAL qpdf syntax (works everywhere):
	//
	// qpdf repaired.pdf output.pdf --pages . "1,3,2" --
	//
	args := []string{
			absRepaired,
			absOut,
			"--pages",
			".",
			order,
			"--",
	}

	log.Println("➡ FINAL qpdf reorder args:", args)

	cmd := exec.Command("qpdf", args...)
	reorderOut, reorderErr := cmd.CombinedOutput()

	if reorderErr != nil {
			log.Println("❌ reorder failed:", reorderErr)
			log.Println("Output:", string(reorderOut))
			return ""
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
