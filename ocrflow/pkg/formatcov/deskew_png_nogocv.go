//go:build nogocv

package formatcov

// DeskewPNGs returns ErrDeskewUnavailable when built without gocv.
// Build with -tags gocv and install OpenCV for real deskew; otherwise
// the caller should use CopyPNGs and set deskewed=false.
func DeskewPNGs(_, _ string) error {
	return ErrDeskewUnavailable
}

func EstimateDeskewAnglePNG(_ string) (float64, error) {
	return 0, ErrDeskewUnavailable
}

func DeskewPNGFileWithAngle(_, _ string, _ float64) error {
	return ErrDeskewUnavailable
}
