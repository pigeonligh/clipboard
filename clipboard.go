package clipboard

type ClipboardItem struct {
	Type string
	Data []byte
}

type ClipboardItems []ClipboardItem

type Clipboard interface {
	Init() error
	GetClipboard() (ClipboardItems, error)
	SetClipboard(ClipboardItems) error
}

var sys Clipboard = defaultClipboard{}

func Init() error {
	return sys.Init()
}

func GetClipboard() (ClipboardItems, error) {
	return sys.GetClipboard()
}

func SetClipboard(items ClipboardItems) error {
	return sys.SetClipboard(items)
}
