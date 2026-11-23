package articles

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArticleListPagination(t *testing.T) {
	asserts := assert.New(t)
	r := setupRouter()

	user := createTestUser()
	author := GetArticleUserModel(user)
	base := time.Now().Add(-time.Hour)
	for i := 0; i < 25; i++ {
		a := ArticleModel{
			Slug:     fmt.Sprintf("page-article-%d-%d", i, time.Now().UnixNano()),
			Title:    fmt.Sprintf("Page Article %d", i),
			Body:     "body",
			Author:   author,
			AuthorID: author.ID,
		}
		a.CreatedAt = base.Add(time.Duration(i) * time.Minute)
		asserts.NoError(SaveOne(&a))
	}

	get := func(offset int) (titles []string, count int) {
		url := fmt.Sprintf("/api/articles?author=%s&limit=20&offset=%d", user.Username, offset)
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		asserts.Equal(http.StatusOK, w.Code)
		var body struct {
			Articles []struct {
				Title string `json:"title"`
			} `json:"articles"`
			ArticlesCount int `json:"articlesCount"`
		}
		asserts.NoError(json.Unmarshal(w.Body.Bytes(), &body))
		for _, a := range body.Articles {
			titles = append(titles, a.Title)
		}
		return titles, body.ArticlesCount
	}

	page1, count1 := get(0)
	asserts.Equal(25, count1, "articlesCount on page 1 is the total")
	if asserts.Len(page1, 20) {
		asserts.Equal("Page Article 24", page1[0], "newest article first")
		asserts.Equal("Page Article 5", page1[19])
	}

	page2, count2 := get(20)
	asserts.Equal(25, count2, "articlesCount on page 2 is the total")
	asserts.Equal([]string{"Page Article 4", "Page Article 3", "Page Article 2", "Page Article 1", "Page Article 0"}, page2)
}
