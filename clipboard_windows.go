//go:build windows

package clipboard

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var win32Format = map[string]uintptr{
	"text/plain": 1,
}

var win32FormatName = map[string]string{
	"text/plain": "Plain Format",
	"text/html":  "HTML Format",
	"text/rtf":   "Rich Text Format",
}

type win32Clipboard struct {
	sync.Mutex
	formats     map[string]uintptr
	formatNames map[uintptr]string
}

func (c *win32Clipboard) registerFormat(name string) (uintptr, error) {
	b, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	ret, err := RegisterClipboardFormatW.DoAndRetrun(uintptr(unsafe.Pointer(b)))
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (c *win32Clipboard) Init() error {
	c.formats = make(map[string]uintptr)
	c.formatNames = make(map[uintptr]string)
	for format, name := range win32FormatName {
		p, found := win32Format[format]
		if !found {
			var err error
			p, err = c.registerFormat(name)
			if err != nil {
				return err
			}
		}
		c.formats[format] = p
		c.formatNames[p] = format
	}
	return nil
}

func (c *win32Clipboard) GetClipboard() (ClipboardItems, error) {
	c.Lock()
	defer c.Unlock()

	ret := make(ClipboardItems, 0)

	err := CallClipboard(func() error {
		cur, _ := EnumClipboardFormats.DoAndRetrun(0)
		for cur != 0 {
			if name, found := c.formatNames[cur]; found {
				h, err := GetClipboardData.DoAndRetrun(cur)
				if err != nil {
					return err
				}
				data, err := getData(syscall.Handle(h))
				if err != nil {
					return err
				}
				ret = append(ret, ClipboardItem{
					Type: name,
					Data: data,
				})
			}
			cur, _ = EnumClipboardFormats.DoAndRetrun(cur)
		}
		return nil
	}, false)

	return ret, err
}

func (c *win32Clipboard) SetClipboard(items ClipboardItems) error {
	c.Lock()
	defer c.Unlock()

	m := make(map[uintptr][]byte)
	for _, item := range items {
		p, found := c.formats[item.Type]
		if !found {
			return fmt.Errorf("clipboard type %v is unsupported", item.Type)
		}
		m[p] = item.Data
	}

	return CallClipboard(func() error {
		for p, data := range m {
			handle, err := allocData(data)
			if err != nil {
				return err
			}
			_, err = SetClipboardData.DoAndRetrun(p, uintptr(handle))
			if err != nil {
				return err
			}
		}
		return nil
	}, true)
}

func init() {
	sys = &win32Clipboard{}
}
