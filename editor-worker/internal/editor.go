package internal

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

// ------------------------------------------------------------
// UNIVERSAL WATERMARK TOOL (TEXT + IMAGE)
// ------------------------------------------------------------

func addWatermark(input string, opts map[string]string) string {
	out := TempName("watermarked", ".pdf")

	wmLayer := createWatermarkLayer(opts)
	if wmLayer == "" {
		log.Println("❌ Failed to create watermark layer")
		return ""
	}

	cmd := exec.Command(
		"convert",
		input,
		wmLayer,
		"-gravity", "center",
		"-compose", "over",
		"-density", "300",
		"-quality", "100",
		"-layers", "optimize",
		out,
	)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Failed to apply watermark:", err)
		return ""
	}

	return out
}

// ------------------------------------------------------------
// CREATE WATERMARK LAYER (TEXT OR IMAGE)
// ------------------------------------------------------------

func createWatermarkLayer(opts map[string]string) string {
	layer := TempName("layer", ".png")

	// Defaults
	wmType := opts["type"]
	if wmType == "" {
		wmType = "text"
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

	position := opts["position"]
	if position == "" {
		position = "center"
	}

	// ------------------------------------------------------------
	// TEXT WATERMARK
	// ------------------------------------------------------------
	if wmType == "text" {
		text := opts["text"]
		if text == "" {
			text = "WATERMARK"
		}

		fontSize := opts["fontSize"]
		if fontSize == "" {
			fontSize = "60"
		}

		cmd := exec.Command(
			"convert",
			"-size", "2480x3508",      // A4 300 DPI
			"xc:none",
			"-gravity", position,
			"-pointsize", fontSize,
			"-fill", color,
			"-annotate", angle, text,
			"-evaluate", "Multiply", opacity,
			layer,
		)

		cmd.Run()
		return applyRepeatIfNeeded(layer, opts)
	}

	// ------------------------------------------------------------
	// IMAGE WATERMARK
	// ------------------------------------------------------------
	if wmType == "image" {
		img := opts["imageUrl"]
		if img == "" {
			log.Println("❌ No image watermark URL provided")
			return ""
		}

		scale := opts["scale"]
		if scale == "" {
			scale = "50"
		}

		sizeVal, _ := strconv.Atoi(scale)

		cmd := exec.Command(
			"convert",
			"-size", "2480x3508", "xc:none",
			img,
			"-resize", fmt.Sprintf("%d%%", sizeVal),
			"-gravity", position,
			"-compose", "over",
			"-composite",
			"-evaluate", "Multiply", opacity,
			layer,
		)

		cmd.Run()
		return applyRepeatIfNeeded(layer, opts)
	}

	return ""
}

// ------------------------------------------------------------
// OPTIONAL: TILE / REPEAT WATERMARK
// ------------------------------------------------------------

func applyRepeatIfNeeded(layer string, opts map[string]string) string {
	if opts["repeat"] != "true" {
		return layer
	}

	tiled := TempName("tiled", ".png")

	cmd := exec.Command(
		"convert",
		layer,
		"-virtual-pixel", "tile",
		"-distort", "AffineProjection", "1,0,0,1,0,0",
		"-write", "mpr:tile",
		"-size", "2480x3508",
		"tile:mpr:tile",
		tiled,
	)

	cmd.Run()
	return tiled
}

// ------------------------------------------------------------
// 2) ADD PAGE NUMBERS
// ------------------------------------------------------------

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

func createPageNumberOverlay() string {
	pdf := TempName("pn", ".pdf")

	cmd := exec.Command(
		"convert",
		"-size", "2480x3508",
		"xc:none",
		"-gravity", "south",
		"-pointsize", "80",
		"label:%Page%",
		pdf,
	)

	cmd.Run()
	return pdf
}

// ------------------------------------------------------------
// 3) ADD HEADER FOOTER
// ------------------------------------------------------------

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
		"xc:none",
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

// ------------------------------------------------------------
// 4) EDIT PDF (simple copy)
// ------------------------------------------------------------

func editPDF(input string, opts map[string]string) string {
	out := TempName("edited", ".pdf")

	cmd := exec.Command("pdftk", input, "output", out)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Edit PDF failed:", err)
		return ""
	}

	return out
}
