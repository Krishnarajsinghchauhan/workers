package internal

import (
	"log"
)

func ProcessJob(job Job) {
	log.Println("⚙ OCR Worker processing:", job.Tool)

	UpdateStatus(job.ID, "processing")

	// 1. Download
	local := DownloadFromS3(job.Files[0])
	if local == "" {
		UpdateStatus(job.ID, "error")
		return
	}

	// ⭐ CRITICAL FIX: remove Unicode/emoji spaces
	local = SafeFilename(local)

	var result string

	// 2. Choose tool
	switch job.Tool {
	case "image-to-text":
		result = extractText(local)

	case "ocr":
		result = runOCR(local)

	case "scanned-enhance":
		result = enhanceScan(local)

	default:
		log.Println("❌ Unknown tool:", job.Tool)
		UpdateStatus(job.ID, "error")
		return
	}

	if result == "" {
		UpdateStatus(job.ID, "error")
		return
	}

	// 3. Upload
	url := UploadToS3(result)

	// 4. Cleanup
	DeleteFile(local)
	DeleteFile(result)

	// 5. Save
	SaveResult(job.ID, url)

	log.Println("✅ OCR job completed:", job.ID)
}
