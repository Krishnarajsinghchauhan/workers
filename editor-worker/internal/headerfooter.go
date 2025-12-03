package internal

import (
	"log"
	"os/exec"
)

func addHeaderFooter(input string, opts map[string]string) string {

	header := opts["header"]
	footer := opts["footer"]

	out := TempName("headerfooter", ".pdf")

	cmd := exec.Command("pdftk", input, "background", createHeaderFooterLayer(header, footer), "output", out)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Header/footer failed:", err)
		return ""
	}

	return out
}

func createHeaderFooterLayer(header, footer string) string {
	out := TempName("hf_layer", ".pdf")

	cmd := exec.Command(
		"convert",
		"-size", "2480x3508",
		"xc:none",
		"-gravity", "north",
		"label:"+header,
		"-gravity", "south",
		"label:"+footer,
		out,
	)

	cmd.Run()
	return out
}
