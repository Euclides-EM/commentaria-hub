package formatcov

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
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

	return envexec.Cmd("bash", "-c", c)
}
