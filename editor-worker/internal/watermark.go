package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)


// -------------------------
// MAIN WATERMARK WRAPPER
// -------------------------
func addWatermark(input string, opts map[string]string) string {
	tempDir := "/tmp/wm_" + RandString()

	// create temp folder
	os.MkdirAll(tempDir, 0755)

	// 1) Create watermark layer (PNG)
	wm := createWatermarkLayer(opts)
	if wm == "" {
		log.Println("❌ Could not create watermark layer")
		return ""
	}

	// 2) Convert PDF → PNG pages
	pagePattern := filepath.Join(tempDir, "page_%03d.png")
	cmd1 := exec.Command("bash", "-c",
		fmt.Sprintf(`convert -density 200 "%s" "%s"`, input, pagePattern))
	if err := cmd1.Run(); err != nil {
		log.Println("❌ Failed PDF → PNG:", err)
		return ""
	}

	// 3) Apply watermark on each page
	outPattern := filepath.Join(tempDir, "wm_%03d.png")
	cmd2 := exec.Command("bash", "-c",
		fmt.Sprintf(`for f in %s/page_*.png; do 
			base=$(basename "$f"); 
			convert "$f" "%s" -gravity center -compose over -composite "%s/wm_$base"; 
		done`,
			tempDir, wm, tempDir,
		))
	if err := cmd2.Run(); err != nil {
		log.Println("❌ Watermark overlay failed:", err)
		return ""
	}

	// 4) Stitch PNG pages → final PDF
	output := TempName("watermarked", ".pdf")
	cmd3 := exec.Command("bash", "-c",
		fmt.Sprintf(`convert "%s/wm_page_*.png" -quality 100 "%s"`,
			tempDir, output))
	if err := cmd3.Run(); err != nil {
		log.Println("❌ Failed to build output PDF:", err)
		return ""
	}

	log.Println("✅ Watermark applied to PDF:", output)
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
