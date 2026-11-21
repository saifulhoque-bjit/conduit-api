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

type articleListBody struct {
	Articles []struct {
		Slug string `json:"slug"`
	} `json:"articles"`
	ArticlesCount int `json:"articlesCount"`
}

func getArticleList(t *testing.T, query string) articleListBody {
	t.Helper()
	r := setupRouter()
	req, _ := http.NewRequest("GET", "/api/articles?"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "GET /api/articles?%s", query)
	var body articleListBody
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	return body
}

// articlesCount is the total number of matches on every page, not only the
// first one, and each page holds the next slice of the newest-first list.
func TestArticleListPageAndTotal(t *testing.T) {
	asserts := assert.New(t)

	suffix := common.RandInt()
	tag := fmt.Sprintf("paging%d", suffix)
	base := time.Now().Add(-time.Hour)
	slugs := make([]string, 25) // slugs[0] is the newest
	for i := 0; i < 25; i++ {
		article, _ := createArticleWithUser(fmt.Sprintf("Paging %d", i), fmt.Sprintf("paging-%d-%d", suffix, i))
		article.CreatedAt = base.Add(time.Duration(i) * time.Minute)
		asserts.NoError(test_db.Model(&article).Update("created_at", article.CreatedAt).Error)
		asserts.NoError(article.setTags([]string{tag}))
		asserts.NoError(SaveOne(&article))
		slugs[24-i] = article.Slug
	}

	// Global Feed: no filter.
	first := getArticleList(t, "limit=10&offset=0")
	asserts.GreaterOrEqual(first.ArticlesCount, 25)
	for _, offset := range []int{10, 20} {
		page := getArticleList(t, fmt.Sprintf("limit=10&offset=%d", offset))
		asserts.Equal(first.ArticlesCount, page.ArticlesCount, "Global Feed articlesCount at offset %d", offset)
	}

	// Filtered to exactly these 25 articles.
	for _, tc := range []struct{ offset, size int }{{0, 10}, {10, 10}, {20, 5}} {
		page := getArticleList(t, fmt.Sprintf("tag=%s&limit=10&offset=%d", tag, tc.offset))
		asserts.Equal(25, page.ArticlesCount, "articlesCount at offset %d", tc.offset)
		got := make([]string, 0, len(page.Articles))
		for _, a := range page.Articles {
			got = append(got, a.Slug)
		}
		asserts.Equal(slugs[tc.offset:tc.offset+tc.size], got, "articles at offset %d", tc.offset)
	}
}
