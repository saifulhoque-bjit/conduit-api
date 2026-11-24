package articles

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/stretchr/testify/assert"
)

type listPage struct {
	Articles []struct {
		Slug string `json:"slug"`
	} `json:"articles"`
	ArticlesCount int `json:"articlesCount"`
}

func getArticleList(t *testing.T, query string) listPage {
	r := setupRouter()
	req, _ := http.NewRequest("GET", "/api/articles"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, query)
	var page listPage
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &page), query)
	return page
}

func slugsOf(page listPage) []string {
	slugs := make([]string, 0, len(page.Articles))
	for _, a := range page.Articles {
		slugs = append(slugs, a.Slug)
	}
	return slugs
}

// SPEC: spec-2bb353
// articlesCount is the total number of matches for every offset and limit,
// and each page is the right newest-first slice.
func TestArticleListCountAcrossPages(t *testing.T) {
	asserts := assert.New(t)

	// Start from an empty article table so the totals are exact.
	test_db.Exec("DELETE FROM article_tags")
	test_db.Unscoped().Where("1 = 1").Delete(&ArticleModel{})

	author := GetArticleUserModel(createTestUser())
	run := common.RandInt()
	tag := fmt.Sprintf("pagetag%d", run)
	base := time.Now().Add(-time.Hour)

	// newest[0] is the newest article. Every third article carries the tag.
	newest := make([]string, 25)
	tagged := make([]string, 0)
	for i := 0; i < 25; i++ {
		slug := fmt.Sprintf("page-%d-%02d", run, i)
		article := ArticleModel{
			Slug:        slug,
			Title:       slug,
			Description: "d",
			Body:        "b",
			Author:      author,
			AuthorID:    author.ID,
		}
		article.CreatedAt = base.Add(time.Duration(i) * time.Minute)
		if i%3 == 0 {
			asserts.NoError(article.setTags([]string{tag}))
		}
		asserts.NoError(SaveOne(&article))
		newest[24-i] = slug
	}
	for _, slug := range newest {
		var n int
		fmt.Sscanf(slug[len(slug)-2:], "%d", &n)
		if n%3 == 0 {
			tagged = append(tagged, slug)
		}
	}

	// No filter, default limit: 20 newest, total 25.
	page := getArticleList(t, "")
	asserts.Equal(newest[:20], slugsOf(page))
	asserts.Equal(25, page.ArticlesCount)

	// 10 per page: pages 1, 2 and 3 of the Home Global Feed.
	for _, c := range []struct {
		offset int
		want   []string
	}{{0, newest[0:10]}, {10, newest[10:20]}, {20, newest[20:25]}} {
		page := getArticleList(t, fmt.Sprintf("?limit=10&offset=%d", c.offset))
		asserts.Equal(c.want, slugsOf(page), "offset=%d", c.offset)
		asserts.Equal(25, page.ArticlesCount, "articlesCount at offset=%d", c.offset)
	}

	// Tag filter: 9 tagged articles, 2 per page.
	asserts.Len(tagged, 9)
	for _, offset := range []int{0, 2, 8} {
		page := getArticleList(t, fmt.Sprintf("?tag=%s&limit=2&offset=%d", tag, offset))
		end := offset + 2
		if end > len(tagged) {
			end = len(tagged)
		}
		asserts.Equal(tagged[offset:end], slugsOf(page), "tag offset=%d", offset)
		asserts.Equal(9, page.ArticlesCount, "tag articlesCount at offset=%d", offset)
	}

	// Invalid limit and offset fall back to 20 from offset 0.
	page = getArticleList(t, "?limit=abc&offset=xyz")
	asserts.Equal(newest[:20], slugsOf(page))
	asserts.Equal(25, page.ArticlesCount)
}
