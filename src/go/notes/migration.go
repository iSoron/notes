package notes

import (
	"encoding/json"
	"github.com/schollz/versionedtext"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

func (site *Site) Migrate() error {
	currentVersion := site.CurrentDataVersion()
	if currentVersion < 2 {
		err := site.MigrateToVersion2()
		if err != nil {
			return err
		}
	}
	if currentVersion < 3 {
		err := site.MigrateToVersion3()
		if err != nil {
			return err
		}
	}
	return nil
}

func (site *Site) CurrentDataVersion() int {
	// If meta.json does not exist, assume version 1
	_, err := os.Stat(path.Join(site.DataDir, "meta.json"))
	if os.IsNotExist(err) {
		return 1
	}

	// Read version from meta.json
	metaFile, err := ioutil.ReadFile(path.Join(site.DataDir, "meta.json"))
	if err != nil {
		log.Fatal(err)
	}
	var meta struct {
		Version int
	}
	err = json.Unmarshal(metaFile, &meta)
	if err != nil {
		log.Fatal(err)
	}
	return meta.Version
}

func (site *Site) MigrateToVersion2() error {
	log.Println("Migrating to Version 2...")

	// Create uploads dir
	err := os.MkdirAll(path.Join(site.DataDir, "uploads"), 0755)
	if err != nil {
		return err
	}

	// Create pages dir
	err = os.MkdirAll(path.Join(site.DataDir, "pages"), 0755)
	if err != nil {
		return err
	}

	// Delete sitemap.xml if it exists
	sitemapPath := path.Join(site.DataDir, "sitemap.xml")
	_, err = os.Stat(sitemapPath)
	if !os.IsNotExist(err) {
		log.Println("Removing: sitemap.xml")
		err = os.Remove(sitemapPath)
		if err != nil {
			return err
		}
	}

	// Traverse files in datadir
	fileInfos, err := ioutil.ReadDir(site.DataDir)
	for _, fileInfo := range fileInfos {
		filePath := path.Join(site.DataDir, fileInfo.Name())

		// Process JSON files (pages)
		if strings.HasSuffix(filePath, ".json") {

			// Read file
			oldPageFile, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			// Parse JSON
			var oldPage struct {
				Name         string
				Text         versionedtext.VersionedText
				RenderedPage string
			}
			err = json.Unmarshal(oldPageFile, &oldPage)
			if err != nil {
				return err
			}

			// Generate new JSON
			var page struct {
				Name            string
				RawContent      versionedtext.VersionedText
				RenderedContent string
			}
			page.Name = oldPage.Name
			page.RawContent = oldPage.Text
			page.RenderedContent = oldPage.RenderedPage
			pageJson, err := json.MarshalIndent(page, "", " ")
			if err != nil {
				return err
			}

			// Write new file
			newFilePath := path.Join(site.DataDir, "pages", page.Name+".json")
			log.Printf("Writing: %s\n", newFilePath)
			err = ioutil.WriteFile(newFilePath, pageJson, 0644)
			if err != nil {
				return err
			}

			// Delete old file
			log.Printf("Removing: %s\n", filePath)
			err = os.Remove(filePath)
			if err != nil {
				return err
			}
		}

		// Process uploads
		if strings.HasSuffix(filePath, ".upload") {
			newFilePath := path.Join(site.DataDir, "uploads", path.Base(filePath))
			log.Printf("Moving: %s -> %s\n", filePath, newFilePath)
			err = os.Rename(filePath, newFilePath)
			if err != nil {
				return err
			}
		}
	}

	// Create new meta JSON
	var meta struct {
		Version int
	}
	meta.Version = 2
	metaJson, err := json.MarshalIndent(meta, "", " ")
	if err != nil {
		return err
	}

	// Write meta.json
	metaPath := path.Join(site.DataDir, "meta.json")
	log.Printf("Writing: %s\n", metaPath)
	err = ioutil.WriteFile(metaPath, metaJson, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (site *Site) MigrateToVersion3() error {
	// Create publish dir
	err := os.MkdirAll(path.Join(site.DataDir, "publish"), 0755)
	if err != nil {
		return err
	}

	// Traverse pages
	fileInfos, err := ioutil.ReadDir(path.Join(site.DataDir, "pages"))
	for _, fileInfo := range fileInfos {
		filePath := path.Join(site.DataDir, "pages", fileInfo.Name())
		// Read file
		oldPageFile, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Parse JSON
		var oldPage struct {
			Name            string
			RawContent      versionedtext.VersionedText
			RenderedContent string
		}
		err = json.Unmarshal(oldPageFile, &oldPage)
		if err != nil {
			return err
		}

		// Generate new JSON
		var page struct {
			Name            string
			ReadOnlyName    string
			RawContent      versionedtext.VersionedText
			RenderedContent string
		}
		page.Name = oldPage.Name
		page.ReadOnlyName = randomString(32)
		page.RawContent = oldPage.RawContent
		page.RenderedContent = oldPage.RenderedContent
		pageJson, err := json.MarshalIndent(page, "", " ")
		if err != nil {
			return err
		}

		// Write new file
		newFilePath := path.Join(site.DataDir, "pages", page.Name+".json")
		log.Printf("Writing: %s\n", filePath)
		err = ioutil.WriteFile(newFilePath, pageJson, 0644)
		if err != nil {
			return err
		}

		// Create symlink
		err = os.Symlink(
			path.Join("../pages", page.Name+".json"),
			path.Join(site.DataDir, "publish", page.ReadOnlyName+".json"),
		)
		if err != nil {
			return err
		}
	}

	// Create new meta JSON
	var meta struct {
		Version int
	}
	meta.Version = 3
	metaJson, err := json.MarshalIndent(meta, "", " ")
	if err != nil {
		return err
	}

	// Write meta.json
	metaPath := path.Join(site.DataDir, "meta.json")
	log.Printf("Writing: %s\n", metaPath)
	err = ioutil.WriteFile(metaPath, metaJson, 0644)
	if err != nil {
		return err
	}

	return nil

}
