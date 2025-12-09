package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/api"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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

	conf := model.NewDefaultConfiguration()

	// Create watermark struct manually
	wm := &model.Watermark{
			Mode:             model.WMImage,
			FileName:         sigFile,
			Pos:              "abs",
			Dx:               float64(x),
			Dy:               float64(y),
			Scale:            1.0,
			Opacity:          1.0,
			Rotation:         0,
			RenderMode:       model.RMFill,
			UserRotation:     0,
			MultiStamp:       false,
			UpdateStamp:      false,
			OnTop:            true,
			Diagonal:         false,
			Page:             nil,
	}

	// Initialize WM, required before applying
	err := wm.Parse()
	if err != nil {
			log.Println("Watermark parse error:", err)
			return pdfFile
	}

	// Apply to PDF
	err = pdfcpu.AddWatermarksFile(pdfFile, output, nil, wm, conf)
	if err != nil {
			log.Println("Apply signature error:", err)
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
