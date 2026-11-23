package articles

// SPEC: spec-aabc69

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/stretchr/testify/assert"
)

// An Article's readingTimeMinutes is its body's word count divided by 200,
// rounded up, and 1 for an empty body.
func TestArticleReadingTimeMinutes(t *testing.T) {
	r := setupRouter()
	cases := []struct {
		words int
		want  int
	}{
		{0, 1},
		{1, 1},
		{200, 1},
		{201, 2},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d words", c.words), func(t *testing.T) {
			article, _ := createArticleWithUser("Reading Time", fmt.Sprintf("reading-time-%d", common.RandInt()))
			article.Body = strings.TrimSpace(strings.Repeat("word ", c.words))
			test_db.Model(&article).Update("body", article.Body)

			req, _ := http.NewRequest("GET", "/api/articles/"+article.Slug, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)

			var got struct {
				Article map[string]interface{} `json:"article"`
			}
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
			assert.Equal(t, float64(c.want), got.Article["readingTimeMinutes"])
		})
	}
}
