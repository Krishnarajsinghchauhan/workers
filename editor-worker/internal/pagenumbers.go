package internal

import (
	"log"
	"os/exec"
)

func addPageNumbers(input string) string {
	out := TempName("pagenumbers", ".pdf")

	cmd := exec.Command("pdftk", input, "background", createPageNumberOverlay(), "output", out)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Page numbers failed:", err)
		return ""
	}

	return out
}

func createPageNumberOverlay() string {
	out := TempName("pn_layer", ".pdf")

	cmd := exec.Command(
		"convert",
		"-size", "2480x3508",
		"xc:none",
		"-gravity", "south",
		"-pointsize", "80",
		"label:%Page%",
		out,
	)

	cmd.Run()
	return out
}
