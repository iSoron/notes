package notes

import (
	"bytes"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var bmPolicy *bluemonday.Policy
var md goldmark.Markdown

func init() {
	rand.Seed(time.Now().Unix())
	bmPolicy = bluemonday.UGCPolicy()
	bmPolicy.RequireNoReferrerOnLinks(true)
	bmPolicy.AllowElements("input")
	bmPolicy.AllowElements("style")
	bmPolicy.AllowAttrs("checked", "disabled", "type").OnElements("input")
	bmPolicy.AllowAttrs("width", "height", "align").OnElements("img")
	bmPolicy.AllowAttrs("style", "class", "align").OnElements("span", "p", "div", "a")
	md = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithExtensions(mathjax.MathJax),
		goldmark.WithExtensions(extension.Typographer),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func contentType(filename string) string {
	switch {
	case strings.Contains(filename, ".css"):
		return "text/css"
	case strings.Contains(filename, ".jpg"):
		return "image/jpeg"
	case strings.Contains(filename, ".png"):
		return "image/png"
	case strings.Contains(filename, ".js"):
		return "application/javascript"
	case strings.Contains(filename, ".xml"):
		return "application/xml"
	}
	return "text/html"
}

func (site *Site) sniffContentType(name string) (string, error) {
	file, err := os.Open(path.Join(site.DataDir, name))
	if err != nil {
		return "", err

	}
	defer file.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	return http.DetectContentType(buffer), nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func MarkdownToHtml(s string) string {
	var output bytes.Buffer
	if err := md.Convert([]byte(s), &output); err != nil {
		panic(err)
	}
	return bmPolicy.SanitizeReader(&output).String()
}
