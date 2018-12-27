package filesystem

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/kildevaeld/strong"

	"github.com/kildevaeld/valse2"
	"github.com/kildevaeld/valse2/httpcontext"
)

type FileSystem struct {
	fs http.FileSystem
}

func (f *FileSystem) Compose(root string) (httpcontext.HandlerFunc, error) {
	group := valse2.NewGroup()
	group.Get("/*filepath", func(ctx *httpcontext.Context) error {
		path := ctx.Params().ByName("filepath")

		file, err := f.fs.Open(path)
		if err != nil {
			fmt.Printf("ERROR %s\n", err)
			return strong.ErrNotFound
		}

		stat, err := file.Stat()
		if err != nil {
			return err
		}

		ext := mime.TypeByExtension(path)
		if ext == "" {
			ext = strong.MIMETextPlain
		}
		ctx.SetBody(file)
		ctx.SetContentType(ext)
		ctx.Header().Set(strong.HeaderContentLength, fmt.Sprintf("%d", stat.Size()))

		return nil
	})
	return group.Compose(root)
}

func New(dir http.FileSystem) valse2.Mountable {
	return &FileSystem{dir}
}
