package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"


)

func resizeSignature(input string, width int) string {
	output := "signature_resized.png"

	cmd := exec.Command("convert", input,
		"-resize", fmt.Sprintf("%dx", width),
		output,
	)

	err := cmd.Run()
	if err != nil {
		log.Println("Signature resize error:", err)
		return input
	}

	return output
}

func addSignatureToPDF(pdfFile, sigFile string, x, y int) string {
	output := "signed.pdf"

	// pdfcpu CLI command (image stamp)
	cmd := exec.Command(
			"pdfcpu", "stamp", "add",
			"-mode", "image",
			"-pos", "abs",
			"-offset", fmt.Sprintf("%d %d", x, y),
			sigFile,
			pdfFile,
			output,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
			log.Println("Signature error:", string(out), err)
			return pdfFile
	}

	return output
}




func ProcessJob(job Job) {
	UpdateStatus(job.ID, "processing")

	// Files
	inputPDF := DownloadFromS3(job.Files[0])
	signatureImg := DownloadFromS3(job.Files[1])

	// Options
	x, _ := strconv.Atoi(job.Options["x"])
	y, _ := strconv.Atoi(job.Options["y"])
	width, _ := strconv.Atoi(job.Options["width"]) // resize signature

	// Step 1: Resize signature
	resized := resizeSignature(signatureImg, width)

	// Step 2: Stamp signature
	output := addSignatureToPDF(inputPDF, resized, x, y)

	// Step 3: Upload completed file
	url := UploadToS3(output)
	SaveResult(job.ID, url)

	// Cleanup
	os.Remove(inputPDF)
	os.Remove(signatureImg)
	os.Remove(resized)
	os.Remove(output)

	UpdateStatus(job.ID, "completed")
}
