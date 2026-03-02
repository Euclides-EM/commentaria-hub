//go:build nogocv

package formatcov

func DenoisePNGs(_, _ string) error {
	return ErrDenoiseUnavailable
}
