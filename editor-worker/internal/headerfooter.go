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
	marginTop := opts["marginTop"]
	if marginTop == "" { marginTop = "80" }
	marginBottom := opts["marginBottom"]
	if marginBottom == "" { marginBottom = "80" }

	// 1. convert pdf → png pages
	pagePattern := filepath.Join(tempDir, "page_%03d.png")
	cmd1 := exec.Command("bash", "-c",
		fmt.Sprintf(`convert -density 200 "%s" "%s"`, input, pagePattern))
	if err := cmd1.Run(); err != nil {
		log.Println("❌ Failed PDF → PNG:", err)
		return ""
	}

	// 2. create header/footer overlay
	layer := filepath.Join(tempDir, "layer.png")
	cmdLayer := exec.Command("bash", "-c",
		fmt.Sprintf(`
convert -size 2480x3508 xc:none \
  -gravity north -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  -gravity south -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  "%s"
`, fontSize, color, marginTop, header, fontSize, color, marginBottom, footer, layer))

	if err := cmdLayer.Run(); err != nil {
		log.Println("❌ Failed to create layer:", err)
		return ""
	}

	// 3. apply layer on each page
	cmd2 := exec.Command("bash", "-c",
		fmt.Sprintf(`
for f in %s/page_*.png; do 
  base=$(basename "$f");
  convert "$f" "%s" -gravity center -compose over -composite "%s/wm_$base";
done
`, tempDir, layer, tempDir))

	if err := cmd2.Run(); err != nil {
		log.Println("❌ Overlay failed:", err)
		return ""
	}

	// 4. rebuild PDF
	output := TempName("headerfooter", ".pdf")
	cmd3 := exec.Command("bash", "-c",
		fmt.Sprintf(`convert "%s/wm_page_*.png" -quality 100 "%s"`, tempDir, output))

	if err := cmd3.Run(); err != nil {
		log.Println("❌ Failed to build final PDF:", err)
		return ""
	}

	return output
}
