package clipboard

import (
	"fmt"

	"golang.design/x/clipboard"
)

var clipboardFormatName = map[clipboard.Format]string{
	clipboard.FmtText:  "text/plain",
	clipboard.FmtImage: "image/png",
}

type defaultClipboard struct{}

func (defaultClipboard) Init() error {
	return clipboard.Init()
}

func (defaultClipboard) GetClipboard() (ClipboardItems, error) {
	ret := ClipboardItems{}
	for _, fmt := range []clipboard.Format{
		clipboard.FmtText,
		clipboard.FmtImage,
	} {
		data := clipboard.Read(fmt)
		if len(data) > 0 {
			ret = append(ret, ClipboardItem{
				Type: clipboardFormatName[fmt],
				Data: data,
			})
		}
	}
	return ret, nil
}

func (defaultClipboard) SetClipboard(items ClipboardItems) error {
	if len(items) == 0 {
		return nil
	}
	if len(items) > 1 {
		return fmt.Errorf("multi-items clipboard is unsupported")
	}
	for _, fmt := range []clipboard.Format{
		clipboard.FmtText,
		clipboard.FmtImage,
	} {
		if clipboardFormatName[fmt] == items[0].Type {
			_ = clipboard.Write(fmt, items[0].Data)
			return nil
		}
	}
	return fmt.Errorf("clipboard type %v is unsupported", items[0].Type)
}
