//go:build windows
// +build windows

package winpty

import (
	"errors"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

func UTF16PtrToString(p *uint16) string {
	var (
		sizeTest uint16
		finalStr = make([]uint16, 0)
	)
	for {
		if *p == uint16(0) {
			break
		}

		finalStr = append(finalStr, *p)
		p = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + unsafe.Sizeof(sizeTest)))
	}
	return string(utf16.Decode(finalStr[0:]))
}

func UTF16PtrFromStringArray(s []string) (*uint16, error) {
	var r []uint16

	for _, ss := range s {
		a, err := syscall.UTF16FromString(ss)
		if err != nil {
			return nil, err
		}

		r = append(r, a...)
	}

	r = append(r, 0)

	return &r[0], nil
}

func GetErrorMessage(err uintptr) string {
	msgPtr, _, _ := winpty_error_msg.Call(err)
	if msgPtr == uintptr(0) {
		return "Unknown Error"
	}
	return UTF16PtrToString((*uint16)(unsafe.Pointer(msgPtr)))
}

func GetErrorCode(err uintptr) uint32 {
	code, _, _ := winpty_error_code.Call(err)
	return uint32(code)
}

func GetExitCodeProcess(procHandle uintptr) (int, error) {
	getExitCode := kernel32.NewProc("GetExitCodeProcess")
	var code uintptr
	ok, _, _ := getExitCode.Call(procHandle, uintptr(unsafe.Pointer(&code)))
	if ok == 0 {
		err := syscall.GetLastError()
		return -1, err
	}
	return int(code), nil
}

func WaitForSingleObject(procHandle uintptr, milliseconds int64) error {
	wait := kernel32.NewProc("WaitForSingleObject")
	res, _, _ := wait.Call(procHandle, uintptr(milliseconds))
	if res == 0 {
		return nil
	} else if res == 258 {
		return errors.New("timeout error")
	} else {
		return errors.New("invalid handle")
	}
}
