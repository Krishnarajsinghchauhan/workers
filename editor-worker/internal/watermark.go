package internal

import (
	"fmt"
	"log"
	"os/exec"
)

// -------------------------
// MAIN WATERMARK WRAPPER
// -------------------------
func addWatermark(input string, opts map[string]string) string {

	output := TempName("watermarked", ".pdf")

	// Create watermark PDF layer
	layerPDF := createWatermarkPDF(opts)
	if layerPDF == "" {
		log.Println("❌ Failed to create watermark layer")
		return ""
	}

	// pdftk multibackground applies watermark on all pages
	cmd := exec.Command("pdftk", input, "multibackground", layerPDF, "output", output)

	if err := cmd.Run(); err != nil {
		log.Println("❌ pdftk overlay error:", err)
		return ""
	}

	return output
}

// -------------------------
// GENERATE THE WATERMARK AS A PDF (NOT PNG)
// -------------------------
func createWatermarkPDF(opts map[string]string) string {
	out := TempName("wm_layer", ".pdf")

	text := opts["text"]
	if text == "" {
		text = "WATERMARK"
	}

	color := opts["color"]
	if color == "" {
		color = "#000000"
	}

	opacity := opts["opacity"]
	if opacity == "" {
		opacity = "0.25"
	}

	angle := opts["angle"]
	if angle == "" {
		angle = "0"
	}

	fontSize := opts["fontSize"]
	if fontSize == "" {
		fontSize = "80"
	}

	position := opts["position"]
	if position == "" {
		position = "center"
	}

	// A4 at 300 DPI = 2480×3508
	cmd := exec.Command(
		"convert",
		"-size", "2480x3508",
		"xc:none",
		"-gravity", position,
		"-pointsize", fontSize,
		"-fill", color,
		"-annotate", angle, text,
		"-alpha", "set",
		"-evaluate", "Multiply", opacity,
		out,
	)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Failed to build watermark layer:", err)
		return ""
	}

	return out
}
