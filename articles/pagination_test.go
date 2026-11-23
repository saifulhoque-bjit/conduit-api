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

type articleListPage struct {
	Articles []struct {
		Slug string `json:"slug"`
	} `json:"articles"`
	ArticlesCount int `json:"articlesCount"`
}

func getArticleListPage(t *testing.T, query string) articleListPage {
	r := setupRouter()
	req, _ := http.NewRequest("GET", "/api/articles?"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var page articleListPage
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &page))
	return page
}

// GET /api/articles pages newest first: every page reports the full total,
// and the second page holds the articles after the first page.
func TestArticleListSecondPage(t *testing.T) {
	asserts := assert.New(t)

	user := createTestUser()
	author := GetArticleUserModel(user)
	suffix := common.RandInt()
	base := time.Now().Add(-time.Hour)
	slugs := make([]string, 3)
	for i := 0; i < 3; i++ {
		slugs[i] = fmt.Sprintf("page-article-%d-%d", suffix, i)
		article := ArticleModel{
			Slug:     slugs[i],
			Title:    slugs[i],
			Body:     "body",
			Author:   author,
			AuthorID: author.ID,
		}
		asserts.NoError(SaveOne(&article))
		test_db.Model(&article).UpdateColumn("created_at", base.Add(time.Duration(i)*time.Minute))
	}

	first := getArticleListPage(t, "author="+user.Username+"&limit=2&offset=0")
	asserts.Equal(3, first.ArticlesCount, "first page total")
	if asserts.Len(first.Articles, 2) {
		asserts.Equal(slugs[2], first.Articles[0].Slug, "newest article first")
		asserts.Equal(slugs[1], first.Articles[1].Slug)
	}

	second := getArticleListPage(t, "author="+user.Username+"&limit=2&offset=2")
	asserts.Equal(3, second.ArticlesCount, "second page total")
	if asserts.Len(second.Articles, 1) {
		asserts.Equal(slugs[0], second.Articles[0].Slug, "oldest article on page two")
	}
}
