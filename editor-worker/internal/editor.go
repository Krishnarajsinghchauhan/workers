package internal

import (
	"log"
	"os/exec"
)

// ----------------------------
// 1) ADD WATERMARK
// ----------------------------
func addWatermark(input string, opts map[string]string) string {
	text := opts["text"]
	if text == "" {
		text = "WATERMARK"
	}

	out := TempName("watermark", ".pdf")

	cmd := exec.Command(
		"pdftk", input, "background", 
		createWatermarkPDF(text),
		"output", out,
	)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Watermark failed:", err)
		return ""
	}

	return out
}

// helper: create temporary watermark layer
func createWatermarkPDF(text string) string {
	pdf := TempName("wm", ".pdf")

	cmd := exec.Command(
		"convert",
		"-size", "2480x3508", // A4
		"-gravity", "center",
		"-fill", "rgba(0,0,0,0.25)",
		"-pointsize", "120",
		"label:"+text,
		pdf,
	)

	cmd.Run()
	return pdf
}

// ----------------------------
// 2) ADD PAGE NUMBERS
// ----------------------------
func addPageNumbers(input string) string {
	out := TempName("pagenumbers", ".pdf")

	cmd := exec.Command(
		"pdftk", input, "background", 
		createPageNumberOverlay(),
		"output", out,
	)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Page numbering failed:", err)
		return ""
	}

	return out
}

// helper: create footer number layer
func createPageNumberOverlay() string {
	pdf := TempName("pn", ".pdf")

	cmd := exec.Command(
		"convert",
		"-size", "2480x3508",
		"-gravity", "south",
		"-pointsize", "80",
		"label:%Page%",
		pdf,
	)

	cmd.Run()
	return pdf
}

// ----------------------------
// 3) ADD HEADER FOOTER
// ----------------------------
func addHeaderFooter(input string, opts map[string]string) string {

	header := opts["header"]
	footer := opts["footer"]

	out := TempName("headerfooter", ".pdf")

	cmd := exec.Command(
		"pdftk", input, "background",
		createHeaderFooterLayer(header, footer),
		"output", out,
	)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Header/Footer failed:", err)
		return ""
	}

	return out
}

func createHeaderFooterLayer(header, footer string) string {
	pdf := TempName("hf", ".pdf")

	cmd := exec.Command(
		"convert",
		"-size", "2480x3508",
		"-gravity", "north",
		"-pointsize", "80",
		"label:"+header,
		"-gravity", "south",
		"label:"+footer,
		pdf,
	)

	cmd.Run()
	return pdf
}

// ----------------------------
// 4) EDIT PDF (simple replace)
// ----------------------------
func editPDF(input string, opts map[string]string) string {

	out := TempName("edited", ".pdf")

	cmd := exec.Command(
		"pdftk", input, "output", out,
	)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Edit PDF failed:", err)
		return ""
	}

	return out
}
