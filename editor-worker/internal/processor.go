package internal

import (
	"log"
)

func ProcessJob(job Job) {
	log.Println("⚙ Editor Worker processing:", job.Tool)

	UpdateStatus(job.ID, "processing")

	local := DownloadFromS3(job.Files[0])
	if local == "" {
		UpdateStatus(job.ID, "error")
		return
	}

	var output string

	switch job.Tool {

	case "watermark":
		output = addWatermark(local, job.Options)

	case "page-numbers":
		output = addPageNumbers(local)

	case "header-footer":
		output = addHeaderFooter(local, job.Options)

	case "edit":
		output = editPDF(local, job.Options)

	default:
		log.Println("❌ Invalid editor tool:", job.Tool)
		UpdateStatus(job.ID, "error")
		return
	}

	url := UploadToS3(output)

	DeleteFile(local)
	DeleteFile(output)

	SaveResult(job.ID, url)
	UpdateStatus(job.ID, "completed")

	log.Println("✅ Editor job done:", job.ID)
}
