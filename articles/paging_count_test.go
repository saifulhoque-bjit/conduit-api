package articles

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/stretchr/testify/assert"
)

// SPEC: spec-dba67e
// articlesCount is the total of all matching articles, counted before offset
// and limit, so the Global Feed can show its page links on every page.
func TestArticleListCountIgnoresOffset(t *testing.T) {
	asserts := assert.New(t)

	suffix := common.RandInt()
	tag := fmt.Sprintf("pagingtag%d", suffix)
	author := createTestUser()
	authorModel := GetArticleUserModel(author)
	fan := createTestUser()
	fanModel := GetArticleUserModel(fan)

	for i := 1; i <= 25; i++ {
		article := ArticleModel{
			Slug:        fmt.Sprintf("paging-%d-%d", suffix, i),
			Title:       fmt.Sprintf("Paging %d %d", suffix, i),
			Description: "d",
			Body:        "b",
			Author:      authorModel,
			AuthorID:    authorModel.ID,
		}
		asserts.NoError(article.setTags([]string{tag}))
		asserts.NoError(SaveOne(&article))
		asserts.NoError(article.favoriteBy(fanModel))
	}

	pageSizes := map[string]int{"0": 10, "10": 10, "20": 5}
	filters := map[string][3]string{
		"tag":       {tag, "", ""},
		"author":    {"", author.Username, ""},
		"favorited": {"", "", fan.Username},
	}
	for name, f := range filters {
		seen := map[uint]bool{}
		for offset, size := range pageSizes {
			articles, count, err := FindManyArticle(f[0], f[1], "10", offset, f[2])
			asserts.NoError(err)
			asserts.Equal(25, count, "%s filter, offset %s: articlesCount", name, offset)
			asserts.Len(articles, size, "%s filter, offset %s: page size", name, offset)
			for _, a := range articles {
				asserts.False(seen[a.ID], "%s filter: article %d on two pages", name, a.ID)
				seen[a.ID] = true
			}
		}
		asserts.Len(seen, 25, "%s filter: pages cover every article", name)
	}

	// The unfiltered Global Feed request, through the router.
	var total int64
	test_db.Model(&ArticleModel{}).Count(&total)
	r := setupRouter()
	for _, offset := range []string{"0", "10", "20"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/articles?limit=10&offset="+offset, nil)
		r.ServeHTTP(w, req)
		asserts.Equal(http.StatusOK, w.Code)
		var body struct {
			ArticlesCount int `json:"articlesCount"`
		}
		asserts.NoError(json.Unmarshal(w.Body.Bytes(), &body))
		asserts.Equal(int(total), body.ArticlesCount, "GET /api/articles offset %s: articlesCount", offset)
	}
}
