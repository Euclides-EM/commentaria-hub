package formatcov

import (
	"fmt"
	"log"
	"os/exec"
)

func Alto2Yolo(altoDir string, outputDir string, shuffle float64, segmontoGranularity string) error {
	args := []string{
		"convert",
		"alto-to-yolo",
		altoDir + "/*.xml",
		outputDir,
	}
	c := "yaltai convert alto-to-yolo " + altoDir + "/*.xml " + outputDir
	if shuffle > 0 {
		args = append(args, "--shuffle", fmt.Sprintf("%.2f", shuffle))
		c += " --shuffle " + fmt.Sprintf("%.2f", shuffle)
	}
	if segmontoGranularity != "" {
		args = append(args, "--segmonto", segmontoGranularity)
		c += " --segmonto " + segmontoGranularity
	}

	cmd := exec.Command("bash", "-c", c)

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("!!! ERROR processing command: yaltai %v", args)
		log.Printf(string(output))
		return fmt.Errorf("failed to convert ALTO to YOLO: %w", err)
	}

	log.Printf("   -> Successfully saved result to: %s", outputDir)
	return nil
}
