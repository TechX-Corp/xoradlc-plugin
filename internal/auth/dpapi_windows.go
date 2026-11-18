//go:build windows

package auth

import (
	"unsafe"

	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"golang.org/x/sys/windows"
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

func dpapi(data []byte, protect bool) ([]byte, error) {
	crypt32 := windows.NewLazySystemDLL("crypt32.dll")
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	var fn *windows.LazyProc
	if protect {
		fn = crypt32.NewProc("CryptProtectData")
	} else {
		fn = crypt32.NewProc("CryptUnprotectData")
	}
	localFree := kernel32.NewProc("LocalFree")

	buf := make([]byte, len(data))
	copy(buf, data)
	src := dataBlob{cbData: uint32(len(buf)), pbData: &buf[0]}
	var dst dataBlob
	const cryptUIForbidden = 0x1
	r, _, _ := fn.Call(
		uintptr(unsafe.Pointer(&src)),
		0, 0, 0, 0,
		uintptr(cryptUIForbidden),
		uintptr(unsafe.Pointer(&dst)),
	)
	if r == 0 {
		return nil, cli.Fail("Could not encrypt / decrypt the token (DPAPI). Run xdlc logout then xdlc login.")
	}
	defer func() {
		if dst.pbData != nil {
			_, _, _ = localFree.Call(uintptr(unsafe.Pointer(dst.pbData)))
		}
	}()
	out := make([]byte, dst.cbData)
	copy(out, unsafe.Slice(dst.pbData, dst.cbData))
	return out, nil
}
