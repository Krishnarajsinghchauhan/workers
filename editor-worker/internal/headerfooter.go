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

	header := escapeText(opts["header"])
	footer := escapeText(opts["footer"])

	fontSize := opts["fontSize"]
	if fontSize == "" { fontSize = "40" }

	color := opts["color"]
	if color == "" { color = "#000000" }

	marginTop := opts["marginTop"]
	if marginTop == "" { marginTop = "80" }

	marginBottom := opts["marginBottom"]
	if marginBottom == "" { marginBottom = "80" }

	// ----------------------------------------
	// 1. Convert PDF → PNG
	// ----------------------------------------
	pagePattern := filepath.Join(tempDir, "page_%03d.png")
	cmd1 := exec.Command("bash", "-c",
		fmt.Sprintf(`convert -density 200 "%s" "%s"`, input, pagePattern))
	if err := cmd1.Run(); err != nil {
		log.Println("❌ PDF → PNG failed:", err)
		return ""
	}

	// ----------------------------------------
	// 2. Read size from first page
	// ----------------------------------------
	identifyCmd := exec.Command("bash", "-c",
		fmt.Sprintf(`identify -format "%%w %%h" "%s/page_001.png"`, tempDir))
	raw, err := identifyCmd.Output()
	if err != nil {
		log.Println("❌ Failed to identify page size:", err)
		return ""
	}
	parts := strings.Split(string(raw), " ")
	width := strings.TrimSpace(parts[0])
	height := strings.TrimSpace(parts[1])

	// ----------------------------------------
	// 3. Build layer with header & footer
	// ----------------------------------------
	layer := filepath.Join(tempDir, "layer.png")

	cmdLayer := exec.Command("bash", "-c",
		fmt.Sprintf(`
convert -size %sx%s xc:none \
  -gravity north -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  -gravity south -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  "%s"
`,
			width, height,
			fontSize, color, marginTop, header,
			fontSize, color, marginBottom, footer,
			layer))

	if err := cmdLayer.Run(); err != nil {
		log.Println("❌ Failed to create layer:", err)
		return ""
	}

	// ----------------------------------------
	// 4. Composite EXACT layer at CENTER ALWAYS
	// ----------------------------------------
	cmd2 := exec.Command("bash", "-c",
		fmt.Sprintf(`
for f in %s/page_*.png; do
  base=$(basename "$f")
  convert "$f" "%s" -gravity center -compose over -composite "%s/out_$base"
done
`, tempDir, layer, tempDir))

	if err := cmd2.Run(); err != nil {
		log.Println("❌ Composite failed:", err)
		return ""
	}

	// ----------------------------------------
	// 5. Rebuild PDF
	// ----------------------------------------
	output := TempName("headerfooter", ".pdf")
	cmd3 := exec.Command("bash", "-c",
		fmt.Sprintf(`convert "%s/out_page_*.png" -quality 95 "%s"`,
			tempDir, output))

	if err := cmd3.Run(); err != nil {
		log.Println("❌ Rebuild failed:", err)
		return ""
	}

	log.Println("✅ Header/Footer applied successfully:", output)
	return output
}

func escapeText(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}


// Escape quotes/newlines in text safely
func escapeText(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
