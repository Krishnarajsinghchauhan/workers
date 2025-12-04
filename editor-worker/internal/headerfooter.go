package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// MAIN PROCESSOR
func addHeaderFooter(input string, opts map[string]string) string {

	tempDir := "/tmp/hf_" + RandString()
	os.MkdirAll(tempDir, 0755)

	header := opts["header"]
	footer := opts["footer"]
	fontSize := opts["fontSize"]
	if fontSize == "" { fontSize = "40" }
	color := opts["color"]
	if color == "" { color = "#000000" }
	align := opts["align"]            // "left", "center", "right"
	marginTop := opts["marginTop"]
	if marginTop == "" { marginTop = "80" }
	marginBottom := opts["marginBottom"]
	if marginBottom == "" { marginBottom = "80" }

	// 1. Convert PDF → PNG pages
	pagePattern := filepath.Join(tempDir, "page_%03d.png")
	cmd1 := exec.Command("bash", "-c",
			fmt.Sprintf(`convert -density 200 "%s" "%s"`, input, pagePattern))

	if err := cmd1.Run(); err != nil {
			log.Println("❌ Failed PDF → PNG:", err)
			return ""
	}

	// 2. Detect width/height of first page
	identifyCmd := exec.Command("bash", "-c",
			fmt.Sprintf(`identify -format "%%w %%h" "%s/page_001.png"`, tempDir))

	out, err := identifyCmd.Output()
	if err != nil {
			log.Println("❌ Failed to identify page size:", err)
			return ""
	}

	parts := strings.Split(string(out), " ")
	width := strings.TrimSpace(parts[0])
	height := strings.TrimSpace(parts[1])

	// Convert alignment → gravity text alignment
	gravity := "center"
	if align == "left" { gravity = "west" }
	if align == "right" { gravity = "east" }

	// 3. Create EXACT-SIZE overlay PNG
	layer := filepath.Join(tempDir, "layer.png")

	cmdLayer := exec.Command("bash", "-c",
			fmt.Sprintf(`
convert -size %sx%s xc:none \
-gravity north -pointsize %s -fill "%s" -annotate +0+%s "%s" \
-gravity south -pointsize %s -fill "%s" -annotate +0+%s "%s" \
"%s"
`, width, height,
					fontSize, color, marginTop, header,
					fontSize, color, marginBottom, footer,
					layer))

	if err := cmdLayer.Run(); err != nil {
			log.Println("❌ Failed to create layer:", err)
			return ""
	}

	// 4. Composite on each page
	cmd2 := exec.Command("bash", "-c",
			fmt.Sprintf(`
for f in %s/page_*.png; do 
base=$(basename "$f");
convert "$f" "%s" -compose over -gravity %s -composite "%s/hf_$base";
done
`, tempDir, layer, gravity, tempDir))

	if err := cmd2.Run(); err != nil {
			log.Println("❌ Overlay failed:", err)
			return ""
	}

	// 5. Rebuild PDF
	output := TempName("headerfooter", ".pdf")
	cmd3 := exec.Command("bash", "-c",
			fmt.Sprintf(`convert "%s/hf_page_*.png" -quality 100 "%s"`,
					tempDir, output))

	if err := cmd3.Run(); err != nil {
			log.Println("❌ Failed to build final PDF:", err)
			return ""
	}

	return output
}

