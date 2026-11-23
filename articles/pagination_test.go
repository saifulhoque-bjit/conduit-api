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

// The list endpoint must return the requested page, newest first, and the
// total number of matching articles regardless of limit and offset.
func TestArticleListPagination(t *testing.T) {
	asserts := assert.New(t)
	r := setupRouter()

	user := createTestUser()
	author := GetArticleUserModel(user)
	base := time.Now().Add(-time.Hour)
	slugs := make([]string, 5) // slugs[0] is the oldest
	for i := range slugs {
		slugs[i] = fmt.Sprintf("page-test-%d-%d", i, common.RandInt())
		article := ArticleModel{
			Slug: slugs[i], Title: slugs[i], Description: "d", Body: "b",
			Author: author, AuthorID: author.ID,
		}
		article.CreatedAt = base.Add(time.Duration(i) * time.Minute)
		asserts.NoError(SaveOne(&article))
	}

	get := func(query string) ([]string, int) {
		req, _ := http.NewRequest("GET", "/api/articles?author="+user.Username+query, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		asserts.Equal(http.StatusOK, w.Code)
		var body struct {
			Articles []struct {
				Slug string `json:"slug"`
			} `json:"articles"`
			ArticlesCount int `json:"articlesCount"`
		}
		asserts.NoError(json.Unmarshal(w.Body.Bytes(), &body))
		got := make([]string, 0, len(body.Articles))
		for _, a := range body.Articles {
			got = append(got, a.Slug)
		}
		return got, body.ArticlesCount
	}

	page1, count1 := get("&limit=2&offset=0")
	asserts.Equal([]string{slugs[4], slugs[3]}, page1, "first page holds the two newest")
	asserts.Equal(5, count1, "articlesCount is the total, not the page size")

	page2, count2 := get("&limit=2&offset=2")
	asserts.Equal([]string{slugs[2], slugs[1]}, page2, "second page holds the next two")
	asserts.Equal(5, count2, "articlesCount is the total on every page")

	page3, count3 := get("&limit=2&offset=4")
	asserts.Equal([]string{slugs[0]}, page3, "last page holds the oldest")
	asserts.Equal(5, count3)
}
