package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func addHeaderFooter(input string, opts map[string]string) string {

	tempDir := "/tmp/hf_" + RandString()
	os.MkdirAll(tempDir, 0755)

	// Extract values
	header := opts["header"]
	footer := opts["footer"]

	fontSize := opts["fontSize"]
	if fontSize == "" {
		fontSize = "40"
	}

	color := opts["color"]
	if color == "" {
		color = "#000000"
	}

	align := opts["align"] // left, center, right

	marginTop := opts["marginTop"]
	if marginTop == "" {
		marginTop = "80"
	}

	marginBottom := opts["marginBottom"]
	if marginBottom == "" {
		marginBottom = "80"
	}

	// -------------------------
	// 1) Convert PDF → PNG
	// -------------------------
	pagePattern := filepath.Join(tempDir, "page_%03d.png")

	cmd1 := exec.Command("bash", "-c",
		fmt.Sprintf(`convert -density 200 "%s" "%s"`, input, pagePattern))

	if err := cmd1.Run(); err != nil {
		log.Println("❌ PDF → PNG failed:", err)
		return ""
	}

	// -------------------------
	// 2) Get size from first page
	// -------------------------
	identifyCmd := exec.Command("bash", "-c",
		fmt.Sprintf(`identify -format "%%w %%h" "%s/page_001.png"`, tempDir))

	raw, err := identifyCmd.Output()
	if err != nil {
		log.Println("❌ Getting size failed:", err)
		return ""
	}

	parts := strings.Split(string(raw), " ")
	width := strings.TrimSpace(parts[0])
	height := strings.TrimSpace(parts[1])

	// -------------------------
	// 3) Alignment → gravity mapping
	// -------------------------
	headerGravity := "north"
	footerGravity := "south"

	textGravity := "center"
	if align == "left" {
		textGravity = "west"
	}
	if align == "right" {
		textGravity = "east"
	}

	// -------------------------
	// 4) Create header + footer layer
	// -------------------------
	layer := filepath.Join(tempDir, "layer.png")

	layerCmd := exec.Command("bash", "-c",
		fmt.Sprintf(`
convert -size %sx%s xc:none \
-gravity %s -pointsize %s -fill "%s" -annotate +0+%s "%s" \
-gravity %s -pointsize %s -fill "%s" -annotate +0+%s "%s" \
"%s"
`,
			width, height,
			headerGravity, fontSize, color, marginTop, escapeText(header),
			footerGravity, fontSize, color, marginBottom, escapeText(footer),
			layer))

	if err := layerCmd.Run(); err != nil {
		log.Println("❌ Layer generation failed:", err)
		return ""
	}

	// -------------------------
	// 5) Composite layer on every page
	// -------------------------
	cmd2 := exec.Command("bash", "-c",
		fmt.Sprintf(`
for f in %s/page_*.png; do
  base=$(basename "$f")
  convert "$f" "%s" -gravity %s -compose over -composite "%s/hf_$base"
done
`, tempDir, layer, textGravity, tempDir))

	if err := cmd2.Run(); err != nil {
		log.Println("❌ Overlay failed:", err)
		return ""
	}

	// -------------------------
	// 6) Rebuild PDF
	// -------------------------
	output := TempName("headerfooter", ".pdf")
	cmd3 := exec.Command("bash", "-c",
		fmt.Sprintf(`convert "%s/hf_page_*.png" -quality 95 "%s"`,
			tempDir, output))

	if err := cmd3.Run(); err != nil {
		log.Println("❌ PDF rebuild failed:", err)
		return ""
	}

	log.Println("✅ Header & Footer applied:", output)
	return output
}

// Escape quotes/newlines in text safely
func escapeText(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
