package notes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jcelliott/lumber"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Site struct {
	DataDir                string
	AllowFileUploads       bool
	MaxUploadSizeInMB      uint
	Logger                 *lumber.ConsoleLogger
	MaxPageContentSizeInMB uint
	SaveMutex              sync.Mutex
}

func (site Site) Run(host string, port string) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"sniffContentType": site.sniffContentType,
	})
	router.LoadHTMLGlob("templates/*")
	router.GET("/", site.GetIndex)
	router.GET("/:page", site.GetPage)
	router.GET("/:page/*command", site.GetPageCommand)
	router.POST("/uploads", site.PostUploads)
	router.POST("/update", site.PostUpdate)
	log.Printf("Listening on %s:%s\n", host, port)
	panic(router.Run(host + ":" + port))
}

func (site *Site) GetIndex(ctx *gin.Context) {
	ctx.Redirect(302, randomString(32)+"/edit")
}

func (site *Site) GetPage(ctx *gin.Context) {
	page := ctx.Param("page")
	if page == "favicon.ico" {
		return
	}
	ctx.Redirect(302, page+"/view")
}

func (site *Site) GetPageCommand(ctx *gin.Context) {
	page := ctx.Param("page")
	command := ctx.Param("command")
	if page == "uploads" {
		site.GetUploads(ctx, command)
		return
	}
	if page == "static" {
		ctx.File(path.Join("static", command))
		return
	}
	if page == "p" {
		site.GetPublished(ctx, command)
		return
	}
	switch command {
	case "/edit":
		site.GetPageEdit(ctx)
	case "/erase":
		site.GetPageErase(ctx)
	case "/view":
		site.GetPageView(ctx)
	case "/raw":
		site.GetPageRaw(ctx)
	case "/history":
		site.GetPageHistory(ctx)
	default:
		return
	}
}

func (site *Site) GetPublished(ctx *gin.Context, readOnlyName string) {
	page, err := site.LoadPublished(readOnlyName)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(page)
	ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
		"ReadPage":        true,
		"RenderedContent": template.HTML(page.RenderedContent),
	})
}

func (site *Site) GetPageView(ctx *gin.Context) {
	pageName := ctx.Param("page")
	page, err := site.Load(pageName)
	if err != nil {
		log.Print(err)
		return
	}

	rawText := page.RawContent.GetCurrent()
	rawHTML := page.RenderedContent

	// Check to see if an old version is requested
	version := ctx.DefaultQuery("version", "invalid")
	versionInt, versionErr := strconv.Atoi(version)
	if versionErr == nil && versionInt > 0 {
		versionText, err := page.RawContent.GetPreviousByTimestamp(int64(versionInt))
		if err == nil {
			rawText = versionText
			rawHTML = MarkdownToHtml(rawText)
		}
	}

	ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
		"ViewPage":        true,
		"Name":            pageName,
		"ReadOnlyName":    page.ReadOnlyName,
		"RenderedContent": template.HTML(rawHTML),
	})
}

func (site *Site) GetPageEdit(ctx *gin.Context) {
	pageName := ctx.Param("page")
	page, err := site.Load(pageName)
	if err != nil {
		log.Print(err)
		return
	}
	ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
		"EditPage":          true,
		"Name":              pageName,
		"ReadOnlyName":      page.ReadOnlyName,
		"RenderedContent":   template.HTML(page.RenderedContent),
		"RawContent":        page.RawContent.GetCurrent(),
		"CurrentUnixTime":   time.Now().Unix(),
		"AllowFileUploads":  site.AllowFileUploads,
		"MaxUploadSizeInMB": site.MaxUploadSizeInMB,
	})
}

func (site *Site) GetPageRaw(ctx *gin.Context) {
	pageName := ctx.Param("page")
	page, err := site.Load(pageName)
	if err != nil {
		log.Print(err)
		return
	}
	ctx.Writer.Header().Set("Content-Type", "text/plain")
	ctx.Data(200, contentType(page.Name), []byte(page.RawContent.GetCurrent()))
}

func (site *Site) GetPageErase(ctx *gin.Context) {
	pageName := ctx.Param("page")
	page, err := site.Load(pageName)
	if err != nil {
		log.Print(err)
		return
	}
	page.Erase()
	ctx.Redirect(302, pageName+"/edit")
}

func (site *Site) GetPageHistory(ctx *gin.Context) {
	pageName := ctx.Param("page")
	page, err := site.Load(pageName)
	if err != nil {
		log.Print(err)
		return
	}
	timestamps, changeSums := page.RawContent.GetMajorSnapshotsAndChangeSums(60)
	n := len(timestamps)
	reversedTimestamps := make([]int64, n)
	reversedChangeSums := make([]int, n)
	reversedFormattedNames := make([]string, n)
	for i, v := range timestamps {
		reversedTimestamps[n-i-1] = timestamps[i]
		reversedChangeSums[n-i-1] = changeSums[i]
		reversedFormattedNames[n-i-1] = time.Unix(v/1000000000, 0).Format("January 2, 2006 15:04:05 MST")
	}
	ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
		"HistoryPage":           true,
		"Name":                  pageName,
		"ReadOnlyName":          page.ReadOnlyName,
		"VersionTimestamps":     reversedTimestamps,
		"VersionFormattedNames": reversedFormattedNames,
		"VersionChangeSums":     reversedChangeSums,
	})
}

func (site *Site) PostUpdate(c *gin.Context) {
	var json struct {
		Page       string `json:"page"`
		RawContent string `json:"new_text"`
		FetchedAt  int64  `json:"fetched_at"`
	}
	err := c.BindJSON(&json)
	if err != nil {
		site.Logger.Trace(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Error",
		})
		return
	}

	// Check content length
	if uint(len(json.RawContent)) > site.MaxPageContentSizeInMB {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Content too long",
		})
		return
	}

	p, err := site.Load(json.Page)
	if err != nil {
		log.Print(err)
		return
	}

	// Check concurrent editing
	if json.FetchedAt > 0 && p.LastEditUnixTime() > json.FetchedAt {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Refusing to overwrite others work",
		})
		return
	}

	err = p.Update(json.RawContent)
	if err != nil {
		log.Print(err)
		return
	}

	err = p.Save()
	if err != nil {
		log.Print(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Saved",
		"unix_time": time.Now().Unix(),
		"rendered":  p.RenderedContent,
	})
}

func (site *Site) PostUploads(c *gin.Context) {
	if !site.AllowFileUploads {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("uploads are disabled on this server"))
		return
	}

	file, info, err := c.Request.FormFile("file")
	defer file.Close()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newName := randomString(64)
	outfile, err := os.Create(path.Join(site.DataDir, "uploads", newName+".upload"))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	file.Seek(0, io.SeekStart)
	_, err = io.Copy(outfile, file)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Location", "/uploads/"+newName+"?filename="+url.QueryEscape(info.Filename))
	return
}

func (site *Site) GetUploads(ctx *gin.Context, filename string) {
	if !strings.HasSuffix(filename, ".upload") {
		filename = filename + ".upload"
	}
	pathname := path.Join(site.DataDir, "uploads", filename)
	ctx.Header("Content-Type", "text/plain")
	ctx.Header(
		"Content-Disposition",
		`attachment; filename="`+ctx.DefaultQuery("filename", "upload")+`"`,
	)
	ctx.File(pathname)
}
