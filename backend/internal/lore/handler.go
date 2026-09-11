package lore

import (
	"embed"
	"io/fs"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed data
var loreFS embed.FS

type Chapter struct {
	Slug    string `json:"slug"`
	Volume  string `json:"volume"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}

func parseTitle(content string) (volume, title string) {
	for _, line := range strings.SplitN(content, "\n", 5) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			raw := strings.TrimPrefix(line, "# ")
			// handles both " --- " and " — " (em dash) separators
			for _, sep := range []string{" --- ", " — ", " – "} {
				if idx := strings.Index(raw, sep); idx >= 0 {
					return strings.TrimSpace(raw[:idx]), strings.TrimSpace(raw[idx+len(sep):])
				}
			}
			return "", raw
		}
	}
	return "", ""
}

func slugFromName(name string) string {
	return strings.TrimSuffix(name, ".md")
}

func listChapters(withContent bool) ([]Chapter, error) {
	entries, err := fs.ReadDir(loreFS, "data")
	if err != nil {
		return nil, err
	}

	var chapters []Chapter
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := e.Name()
		// skip meta/non-chapter files
		skip := map[string]bool{
			"README.md": true, "FONTES_OFICIAIS.md": true,
			"00_README.md": true, "FONTES.md": true,
			"NOTAS_DE_CONTINUIDADE.md": true,
		}
		if skip[name] {
			continue
		}

		raw, err := loreFS.ReadFile("data/" + name)
		if err != nil {
			continue
		}
		content := string(raw)
		volume, title := parseTitle(content)

		ch := Chapter{
			Slug:   slugFromName(name),
			Volume: volume,
			Title:  title,
		}
		if withContent {
			ch.Content = content
		}
		chapters = append(chapters, ch)
	}

	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Slug < chapters[j].Slug
	})
	return chapters, nil
}

func ListChapters(c *gin.Context) {
	full := c.Query("full") == "true"
	chapters, err := listChapters(full)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chapters)
}

func GetChapter(c *gin.Context) {
	slug := c.Param("slug")
	if strings.Contains(slug, "/") || strings.Contains(slug, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug"})
		return
	}

	raw, err := loreFS.ReadFile("data/" + slug + ".md")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}
	content := string(raw)
	volume, title := parseTitle(content)

	c.JSON(http.StatusOK, Chapter{
		Slug:    slug,
		Volume:  volume,
		Title:   title,
		Content: content,
	})
}

// ── Card showcases ────────────────────────────────────────────────

type CardShowcase struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}

func listShowcaseData(withContent bool) ([]CardShowcase, error) {
	entries, err := fs.ReadDir(loreFS, "data/cards")
	if err != nil {
		return []CardShowcase{}, nil
	}
	var out []CardShowcase
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := loreFS.ReadFile("data/cards/" + e.Name())
		if err != nil {
			continue
		}
		content := string(raw)
		_, title := parseTitle(content)
		if title == "" {
			title = strings.TrimSuffix(e.Name(), ".md")
		}
		cs := CardShowcase{Slug: strings.TrimSuffix(e.Name(), ".md"), Title: title}
		if withContent {
			cs.Content = content
		}
		out = append(out, cs)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out, nil
}

func ListCardShowcases(c *gin.Context) {
	full := c.Query("full") == "true"
	showcases, err := listShowcaseData(full)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, showcases)
}

func GetCardShowcase(c *gin.Context) {
	slug := c.Param("slug")
	if strings.Contains(slug, "/") || strings.Contains(slug, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug"})
		return
	}
	raw, err := loreFS.ReadFile("data/cards/" + slug + ".md")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "showcase not found"})
		return
	}
	content := string(raw)
	_, title := parseTitle(content)
	c.JSON(http.StatusOK, CardShowcase{Slug: slug, Title: title, Content: content})
}

func RegisterRoutes(r gin.IRouter) {
	r.GET("/lore", ListChapters)
	r.GET("/lore/:slug", GetChapter)
	r.GET("/lore-cards", ListCardShowcases)
	r.GET("/lore-cards/:slug", GetCardShowcase)
}
