//go:build windows

package clipboard

import (
	"syscall"
	"unsafe"

	"github.com/chai2010/cgo"
)

type DLLProc struct {
	*syscall.LazyProc
}

func (c *DLLProc) Do(args ...uintptr) error {
	_, _, err := c.LazyProc.Call(args...)
	if err != nil && err != syscall.Errno(0) {
		return err
	}
	return nil
}

func (c *DLLProc) DoAndRetrun(args ...uintptr) (uintptr, error) {
	ret, _, err := c.LazyProc.Call(args...)
	if err != nil && err != syscall.Errno(0) {
		return 0, err
	}
	return ret, nil
}

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	RegisterClipboardFormatW = &DLLProc{user32.NewProc("RegisterClipboardFormatW")}
	OpenClipboard            = &DLLProc{user32.NewProc("OpenClipboard")}
	EmptyClipboard           = &DLLProc{user32.NewProc("EmptyClipboard")}
	CloseClipboard           = &DLLProc{user32.NewProc("CloseClipboard")}
	SetClipboardData         = &DLLProc{user32.NewProc("SetClipboardData")}
	GetClipboardData         = &DLLProc{user32.NewProc("GetClipboardData")}
	EnumClipboardFormats     = &DLLProc{user32.NewProc("EnumClipboardFormats")}
	GetClipboardFormatNameW  = &DLLProc{user32.NewProc("GetClipboardFormatNameW")}
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	globalAlloc  = &DLLProc{kernel32.NewProc("GlobalAlloc")}
	globalLock   = &DLLProc{kernel32.NewProc("GlobalLock")}
	globalUnlock = &DLLProc{kernel32.NewProc("GlobalUnlock")}
)

func CallClipboard(f func() error, reset bool) error {
	err := OpenClipboard.Do(0)
	if err != nil {
		return err
	}
	defer func() {
		_ = CloseClipboard.Do()
	}()

	if reset {
		err = EmptyClipboard.Do()
		if err != nil {
			return err
		}
	}

	return f()
}

func allocData(data []byte) (syscall.Handle, error) {
	handle, err := globalAlloc.DoAndRetrun(
		0x02, // GMEM_MOVEABLE
		uintptr(len(data)+1),
	)
	if err != nil {
		return 0, err
	}

	ptr, err := globalLock.DoAndRetrun(handle)
	if err != nil {
		return 0, err
	}
	cgo.VoidPointer(ptr).Memcpy(cgo.VoidPointer(unsafe.Pointer(&data[0])), len(data))
	err = globalUnlock.Do(handle)
	if err != nil {
		return 0, err
	}

	return syscall.Handle(handle), nil
}

func getData(handle syscall.Handle) ([]byte, error) {
	ptr, err := globalLock.DoAndRetrun(uintptr(handle))
	if err != nil {
		return nil, err
	}
	s := cgo.VoidPointer(ptr).GoString()
	_ = globalUnlock.Do(uintptr(handle))
	return []byte(s), nil
}
