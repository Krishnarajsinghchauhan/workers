package internal

import (
	"log"
	"os/exec"
)

// ----------------------------
// Convert Images → PDF
// ----------------------------
func imagesToPDF(input string) string {
	output := TempFile("images_to_pdf", ".pdf")

	cmd := exec.Command("convert", input, output)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Image to PDF failed:", err)
		return ""
	}

	return output
}

// ----------------------------
// Convert PDF → Images
// ----------------------------
func pdfToImages(input string) string {
	output := TempFile("pdf_to_image", ".png")

	cmd := exec.Command("convert", input, output)

	if err := cmd.Run(); err != nil {
		log.Println("❌ PDF to image failed:", err)
		return ""
	}

	return output
}

// ----------------------------
// MAIN JOB PROCESSOR
// ----------------------------
func ProcessJob(job Job) {
	log.Println("⚙ Processing Image job:", job.Tool)

	UpdateStatus(job.ID, "processing")

	// Download file from S3
	local := DownloadFromS3(job.Files[0])
	if local == "" {
		UpdateStatus(job.ID, "error")
		return
	}

	var output string

	switch job.Tool {

	case "jpg-to-pdf", "png-to-pdf":
		output = imagesToPDF(local)

	case "pdf-to-jpg", "pdf-to-png":
		output = pdfToImages(local)

	default:
		log.Println("❌ Unknown image tool:", job.Tool)
		UpdateStatus(job.ID, "error")
		return
	}

	if output == "" {
		UpdateStatus(job.ID, "error")
		return
	}

	// Upload output to S3
	finalURL := UploadToS3(output)
	SaveResult(job.ID, finalURL)

	// Cleanup
	DeleteFile(local)
	DeleteFile(output)

	log.Println("✅ Image job completed:", job.ID)
}
