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

	log.Println("🔵 Received Header/Footer Options:", opts)

	tempDir := "/tmp/hf_" + RandString()
	os.MkdirAll(tempDir, 0755)

	header := escapeText(opts["header"])
	footer := escapeText(opts["footer"])

	fontSize := opts["fontSize"]
	if fontSize == "" {
		fontSize = "40"
	}

	color := opts["color"]
	if color == "" {
		color = "#000000"
	}

	marginTop := opts["marginTop"]
	if marginTop == "" {
		marginTop = "80"
	}

	marginBottom := opts["marginBottom"]
	if marginBottom == "" {
		marginBottom = "80"
	}

	log.Println("✨ Parsed Options:",
		"header=", header,
		"footer=", footer,
		"fontSize=", fontSize,
		"color=", color,
		"marginTop=", marginTop,
		"marginBottom=", marginBottom,
	)

	// -------------------------
	// 1) PDF → PNG
	// -------------------------
	pagePattern := filepath.Join(tempDir, "page_%03d.png")
	cmd1 := fmt.Sprintf(`convert -density 200 "%s" "%s"`, input, pagePattern)

	log.Println("🟡 Running:", cmd1)

	if err := exec.Command("bash", "-c", cmd1).Run(); err != nil {
		log.Println("❌ PDF→PNG failed:", err)
		return ""
	}

	// -------------------------
	// 2) Identify size
	// -------------------------
	identifyCmd := fmt.Sprintf(`identify -format "%%w %%h" "%s/page_001.png"`, tempDir)

	sizeOut, err := exec.Command("bash", "-c", identifyCmd).Output()
	if err != nil {
		log.Println("❌ Identify failed:", err)
		return ""
	}

	parts := strings.Split(string(sizeOut), " ")
	width := strings.TrimSpace(parts[0])
	height := strings.TrimSpace(parts[1])

	log.Println("📏 Page Size:", width, "x", height)

	// -------------------------
	// 3) Create Layer
	// -------------------------
	layer := filepath.Join(tempDir, "layer.png")

	layerCmd := fmt.Sprintf(`
convert -size %sx%s xc:none \
  -gravity north -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  -gravity south -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  "%s"
`,
		width, height,
		fontSize, color, marginTop, header,
		fontSize, color, marginBottom, footer,
		layer,
	)

	log.Println("🟡 Creating Layer:", layerCmd)

	if err := exec.Command("bash", "-c", layerCmd).Run(); err != nil {
		log.Println("❌ Layer creation failed:", err)
		return ""
	}

	// -------------------------
	// 4) Apply Layer
	// -------------------------
	compositeCmd := fmt.Sprintf(`
for f in %s/page_*.png; do
  base=$(basename "$f")
  convert "$f" "%s" -compose over -composite "%s/out_$base"
done
`, tempDir, layer, tempDir)

	log.Println("🟡 Compositing:", compositeCmd)

	if err := exec.Command("bash", "-c", compositeCmd).Run(); err != nil {
		log.Println("❌ Composite failed:", err)
		return ""
	}

	// -------------------------
	// 5) Rebuild PDF
	// -------------------------
	output := TempName("headerfooter", ".pdf")

	rebuildCmd := fmt.Sprintf(`convert "%s/out_page_*.png" -quality 95 "%s"`, tempDir, output)

	log.Println("🟡 Rebuilding PDF:", rebuildCmd)

	if err := exec.Command("bash", "-c", rebuildCmd).Run(); err != nil {
		log.Println("❌ Rebuild failed:", err)
		return ""
	}

	log.Println("✅ FINAL PDF GENERATED:", output)
	return output
}

func escapeText(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
