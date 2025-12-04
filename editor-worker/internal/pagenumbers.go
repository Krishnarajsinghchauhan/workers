package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// ----------------------------
// ADD PAGE NUMBERS TO PDF
// ----------------------------
func addPageNumbers(input string) string {
	tempDir := "/tmp/pn_" + RandString()
	os.MkdirAll(tempDir, 0755)

	// Convert PDF → PNG pages
	pagePattern := filepath.Join(tempDir, "page_%03d.png")
	cmd1 := exec.Command(
		"bash",
		"-c",
		fmt.Sprintf(`convert -density 150 "%s" "%s"`, input, pagePattern),
	)

	if err := cmd1.Run(); err != nil {
		log.Println("❌ Failed PDF → PNG:", err)
		return ""
	}

	// Apply page number on each page
	cmd2 := exec.Command("bash", "-c",
		fmt.Sprintf(`
		i=1
		for f in %s/page_*.png; do
			base=$(basename "$f")
			convert "$f" \
				-gravity south \
				-pointsize 50 \
				-fill black \
				-annotate +0+30 "Page $i" \
				"%s/pn_$base"
			i=$((i+1))
		done
	`, tempDir, tempDir))

	if err := cmd2.Run(); err != nil {
		log.Println("❌ Page number overlay failed:", err)
		return ""
	}

	// Combine back to PDF
	output := TempName("page_numbers", ".pdf")

	cmd3 := exec.Command(
		"bash",
		"-c",
		fmt.Sprintf(`convert "%s/pn_page_*.png" -quality 100 "%s"`, tempDir, output),
	)

	if err := cmd3.Run(); err != nil {
		log.Println("❌ Failed to build output PDF:", err)
		return ""
	}

	log.Println("✅ Page numbers added:", output)
	return output
}
