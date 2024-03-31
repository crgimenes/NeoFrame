package screen

var (
	imbBuf        []byte
	width, height int
)

func init() {
	width, height = GetScreenSize()
	imbBuf = make([]byte, width*height*4)
}
