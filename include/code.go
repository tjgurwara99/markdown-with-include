package include

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
)

func includeRenderHook(root fs.FS, currentPath string) html.RenderNodeFunc {
	return func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
		if link, ok := node.(*ast.Link); ok && entering {
			if len(link.Container.Children) != 1 {
				return ast.GoToNext, false
			}
			child := link.Container.Children[0]
			textNode, ok := child.(*ast.Text)
			if !ok {
				return ast.GoToNext, false
			}
			if !bytes.Equal(textNode.Literal, []byte("!include")) {
				return ast.GoToNext, false
			}
			if bytes.HasPrefix(link.Destination, []byte(".")) {
				link.Destination = []byte(filepath.Join(currentPath, string(link.Destination)))
			}
			file, err := root.Open(strings.TrimPrefix(string(link.Destination), "/"))
			if err != nil {
				log.Println("failed to open file", err)
				return ast.Terminate, false
			}
			defer file.Close()
			filecontent, err := io.ReadAll(file)
			if err != nil {
				return ast.Terminate, false
			}
			html := markdown.ToHTML(filecontent, nil, nil)
			fmt.Fprintf(w, "%s\n", html)
			return ast.SkipChildren, true
		}
		return ast.GoToNext, false
	}
}

func newIncludeRenderer(root fs.FS, currentPath string) *html.Renderer {
	opts := html.RendererOptions{
		Flags:          html.CommonFlags,
		RenderNodeHook: includeRenderHook(root, currentPath),
	}
	return html.NewRenderer(opts)
}

func Render(root fs.FS, base string, data []byte) []byte {
	renderer := newIncludeRenderer(root, base)
	return markdown.ToHTML(data, nil, renderer)
}

func FileServer(root fs.FS) http.Handler {
	return &fileHandler{handler: http.FileServerFS(root), fs: root}
}

type fileHandler struct {
	handler http.Handler
	fs      fs.FS
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	if !strings.HasSuffix(upath, ".md") {
		f.handler.ServeHTTP(w, r)
		return
	}
	file, err := f.fs.Open(strings.TrimPrefix(path.Clean(upath), "/"))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	rendered := Render(f.fs, filepath.Dir(upath), content)
	fmt.Fprintf(w, "%s\n", rendered)
}
