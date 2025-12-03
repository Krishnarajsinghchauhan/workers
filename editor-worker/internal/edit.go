package internal

import (
	"log"
	"os/exec"
)

func editPDF(input string, opts map[string]string) string {
	out := TempName("edited", ".pdf")

	cmd := exec.Command("pdftk", input, "output", out)

	if err := cmd.Run(); err != nil {
		log.Println("❌ Edit failed:", err)
		return ""
	}

	return out
}
