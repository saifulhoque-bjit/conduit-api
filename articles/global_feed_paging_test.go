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

// Global Feed: GET /api/articles?limit&offset must report the total number of
// articles in articlesCount, not the size of the returned page.
func TestGlobalFeedPagingCount(t *testing.T) {
	asserts := assert.New(t)
	r := setupRouter()

	for i := 0; i < 3; i++ {
		createArticleWithUser("Paging Article", fmt.Sprintf("paging-article-%d-%d", i, common.RandInt()))
	}
	var total int64
	test_db.Model(&ArticleModel{}).Count(&total)

	type page struct {
		Articles      []map[string]interface{} `json:"articles"`
		ArticlesCount int                      `json:"articlesCount"`
	}
	get := func(url string) page {
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		asserts.Equal(http.StatusOK, w.Code)
		var p page
		asserts.NoError(json.Unmarshal(w.Body.Bytes(), &p))
		return p
	}

	first := get("/api/articles?limit=2&offset=0")
	asserts.Len(first.Articles, 2)
	asserts.Equal(int(total), first.ArticlesCount, "first page must report the total")

	second := get("/api/articles?limit=2&offset=2")
	asserts.Equal(int(total), second.ArticlesCount, "second page must report the total")
}
