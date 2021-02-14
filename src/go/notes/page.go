package notes

import (
	"encoding/json"
	"fmt"
	"github.com/schollz/versionedtext"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

type Page struct {
	Site            *Site
	Name            string
	ReadOnlyName    string
	RawContent      versionedtext.VersionedText
	RenderedContent string
}

func (p Page) LastEditUnixTime() int64 {
	return p.RawContent.LastEditTime() / 1000000000
}

func (site *Site) Load(name string) (*Page, error) {
	matched, _ := regexp.MatchString(`^[a-z0-9]*$`, name)
	if !matched || len(name) > 64 || len(name) < 3 {
		return nil, fmt.Errorf("Invalid note name: %s", name)
	}
	page := new(Page)
	page.Site = site
	page.Name = name
	page.ReadOnlyName = randomString(32)
	page.RawContent = versionedtext.NewVersionedText("")
	page.Render()

	// Read file
	pageJson, err := ioutil.ReadFile(page.FileName())
	if os.IsNotExist(err) {
		// If page does not exist, create a new one
		err = os.Symlink(
			path.Join("../pages", page.Name + ".json"),
			path.Join(site.DataDir, "publish", page.ReadOnlyName + ".json"),
		)
		if err != nil {
			return nil, err
		}
		return page, nil
	} else if err != nil {
		// Throw any other errors
		return nil, err
	}

	// Parse JSON file
	err = json.Unmarshal(pageJson, &page)
	if err != nil {
		return nil, err
	}
	return page, err
}

func (site *Site) LoadPublished(readOnlyName string) (*Page, error) {
	page := new(Page)
	filename := path.Join(site.DataDir, "publish", readOnlyName + ".json")

	// Read file
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Parse JSON file
	err = json.Unmarshal(fileContents, &page)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (p *Page) Update(newRawContent string) error {
	newRawContent = strings.TrimRight(newRawContent, "\n\t ")
	p.RawContent.Update(newRawContent)
	p.Render()
	return p.Save()
}

func (p *Page) Render() {
	p.RenderedContent = MarkdownToHtml(p.RawContent.GetCurrent())
}

func (p *Page) Save() error {
	p.Site.SaveMutex.Lock()
	defer p.Site.SaveMutex.Unlock()
	var tmp struct {
		Name            string
		ReadOnlyName    string
		RawContent      versionedtext.VersionedText
		RenderedContent string
	}
	tmp.Name = p.Name
	tmp.RawContent = p.RawContent
	tmp.RenderedContent = p.RenderedContent
	tmp.ReadOnlyName = p.ReadOnlyName
	bJSON, err := json.MarshalIndent(tmp, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p.FileName(), bJSON, 0644)
}

func (p *Page) Erase() error {
	p.Site.Logger.Trace("Erasing " + p.Name)
	return os.Remove(p.FileName())
}

func (p *Page) FileName() string {
	return path.Join(p.Site.DataDir, "pages", p.Name+".json")
}
