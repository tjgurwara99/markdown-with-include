package include

import (
	"embed"
	"testing"
)

//go:embed random.md
var root embed.FS

func TestMarkdown(t *testing.T) {
	document := `# Testing doc

There is something here.

Now using include:

[!include](/random.md)

[link](https://something.com)
`
	doc := Render(root, ".", []byte(document))
	t.Log(doc)
}
