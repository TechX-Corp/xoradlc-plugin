//go:build !windows

package auth

func dpapi(data []byte, protect bool) ([]byte, error) {
	_ = protect
	return data, nil
}