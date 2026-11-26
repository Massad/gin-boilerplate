package article

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	validate "github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeRepository stands in for the repository port: it records the forwarded
// calls and replays canned results or errors. A zero-row update or delete is
// modelled by an ErrNotFound answer, the contract the real repository keeps.
type fakeRepository struct {
	createID     uuid.UUID
	createErr    error
	oneArticle   ArticleResponse
	oneErr       error
	allResults   []Result
	allErr       error
	updateErr    error
	deleteErr    error
	createCalls  int
	createForm   CreateArticleForm
	createTenant uuid.UUID
	createUser   uuid.UUID
	oneCalls     int
	oneTenant    uuid.UUID
	oneUser      uuid.UUID
	oneID        uuid.UUID
	allCalls     int
	allTenant    uuid.UUID
	allUser      uuid.UUID
	updateCalls  int
	updateTenant uuid.UUID
	updateUser   uuid.UUID
	updateID     uuid.UUID
	updateFormID uuid.UUID
	updateForm   CreateArticleForm
	deleteCalls  int
	deleteTenant uuid.UUID
	deleteUser   uuid.UUID
	deleteID     uuid.UUID
}

func (f *fakeRepository) Create(_ context.Context, tenantID, userID uuid.UUID, form CreateArticleForm) (uuid.UUID, error) {
	f.createCalls++
	f.createTenant, f.createUser, f.createForm = tenantID, userID, form
	return f.createID, f.createErr
}

func (f *fakeRepository) One(_ context.Context, tenantID, userID, id uuid.UUID) (ArticleResponse, error) {
	f.oneCalls++
	f.oneTenant, f.oneUser, f.oneID = tenantID, userID, id
	return f.oneArticle, f.oneErr
}

func (f *fakeRepository) All(_ context.Context, tenantID, userID uuid.UUID) ([]Result, error) {
	f.allCalls++
	f.allTenant, f.allUser = tenantID, userID
	return f.allResults, f.allErr
}

func (f *fakeRepository) Update(_ context.Context, tenantID, userID, id uuid.UUID, form CreateArticleForm) error {
	f.updateCalls++
	f.updateTenant, f.updateUser, f.updateID, f.updateForm = tenantID, userID, id, form
	return f.updateErr
}

func (f *fakeRepository) Delete(_ context.Context, tenantID, userID, id uuid.UUID) error {
	f.deleteCalls++
	f.deleteTenant, f.deleteUser, f.deleteID = tenantID, userID, id
	return f.deleteErr
}

const (
	identityUserID   = "11111111-1111-1111-1111-111111111111"
	identityTenantID = "22222222-2222-2222-2222-222222222222"
)

var (
	identityUser   = uuid.MustParse(identityUserID)
	identityTenant = uuid.MustParse(identityTenantID)
)

// articleID is a well-formed article route parameter.
var articleID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

// newTestRouter wires the article handler on a Gin engine shaped like the
// production /v1 group, with a middleware standing in for TokenAuth so the
// identity helpers read the documented context keys.
func newTestRouter(fake *fakeRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	binding.Validator = new(validate.DefaultValidator)

	h := NewHandler(NewService(fake))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", identityUserID)
		c.Set("tenantID", identityTenantID)
		c.Next()
	})
	r.POST("/v1/article", h.Create)
	r.GET("/v1/articles", h.All)
	r.GET("/v1/article/:id", h.One)
	r.PUT("/v1/article/:id", h.Update)
	r.DELETE("/v1/article/:id", h.Delete)
	return r
}

// doArticle runs one request and returns the recorder.
func doArticle(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, path, nil)
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		req = httptest.NewRequest(method, path, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func TestCreateArticle(t *testing.T) {
	caseForm := CreateArticleForm{Title: "Testing article title", Content: "Testing article content"}

	t.Run("created", func(t *testing.T) {
		fake := &fakeRepository{createID: articleID}
		resp := doArticle(newTestRouter(fake), http.MethodPost, "/v1/article", caseForm)

		require.Equal(t, http.StatusOK, resp.Code)
		require.JSONEq(t, `{"message":"Article created","id":"33333333-3333-3333-3333-333333333333"}`, resp.Body.String())
		require.Equal(t, 1, fake.createCalls)
		require.Equal(t, identityTenant, fake.createTenant, "tenant id comes from the TokenAuth context key")
		require.Equal(t, identityUser, fake.createUser, "user id comes from the TokenAuth context key")
		require.Equal(t, caseForm, fake.createForm)
	})

	t.Run("repository error", func(t *testing.T) {
		fake := &fakeRepository{createErr: errors.New("boom")}
		resp := doArticle(newTestRouter(fake), http.MethodPost, "/v1/article", caseForm)

		require.Equal(t, http.StatusNotAcceptable, resp.Code)
		require.JSONEq(t, `{"message":"Article could not be created"}`, resp.Body.String())
	})

	t.Run("binding error", func(t *testing.T) {
		fake := &fakeRepository{}
		resp := doArticle(newTestRouter(fake), http.MethodPost, "/v1/article", CreateArticleForm{Title: "ab", Content: "x"})

		require.Equal(t, http.StatusNotAcceptable, resp.Code)
		require.JSONEq(t, `{"message":"Title should be between 3 to 100 characters"}`, resp.Body.String())
		require.Zero(t, fake.createCalls)
	})
}

func TestListArticles(t *testing.T) {
	list := []Result{{Data: []ArticleResponse{{ID: articleID.String(), Title: "Testing article title"}}, Meta: Meta{Total: 1}}}

	t.Run("listed", func(t *testing.T) {
		fake := &fakeRepository{allResults: list}
		resp := doArticle(newTestRouter(fake), http.MethodGet, "/v1/articles", nil)

		require.Equal(t, http.StatusOK, resp.Code)
		var got struct {
			Results []Result `json:"results"`
		}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
		require.Equal(t, list, got.Results)
	})

	t.Run("repository error", func(t *testing.T) {
		fake := &fakeRepository{allErr: errors.New("boom")}
		resp := doArticle(newTestRouter(fake), http.MethodGet, "/v1/articles", nil)

		require.Equal(t, http.StatusNotAcceptable, resp.Code)
		require.JSONEq(t, `{"Message":"Could not get articles"}`, resp.Body.String())
	})
}

func TestGetArticle(t *testing.T) {
	found := ArticleResponse{ID: articleID.String(), Title: "Testing article title"}

	t.Run("found", func(t *testing.T) {
		fake := &fakeRepository{oneArticle: found}
		resp := doArticle(newTestRouter(fake), http.MethodGet, "/v1/article/"+articleID.String(), nil)

		require.Equal(t, http.StatusOK, resp.Code)
		var got struct {
			Data ArticleResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
		require.Equal(t, found, got.Data)
		require.Equal(t, identityTenant, fake.oneTenant)
		require.Equal(t, identityUser, fake.oneUser)
		require.Equal(t, articleID, fake.oneID)
	})

	t.Run("repository error maps to 404", func(t *testing.T) {
		fake := &fakeRepository{oneErr: errors.New("boom")}
		resp := doArticle(newTestRouter(fake), http.MethodGet, "/v1/article/"+articleID.String(), nil)

		require.Equal(t, http.StatusNotFound, resp.Code)
		require.JSONEq(t, `{"Message":"Article not found"}`, resp.Body.String())
	})

	t.Run("invalid parameter", func(t *testing.T) {
		fake := &fakeRepository{}
		resp := doArticle(newTestRouter(fake), http.MethodGet, "/v1/article/invalid", nil)

		require.Equal(t, http.StatusNotFound, resp.Code)
		require.JSONEq(t, `{"Message":"Invalid parameter"}`, resp.Body.String())
		require.Zero(t, fake.oneCalls)
	})
}

func TestUpdateArticle(t *testing.T) {
	caseForm := CreateArticleForm{Title: "Testing new article title", Content: "Testing new article content"}

	t.Run("updated", func(t *testing.T) {
		fake := &fakeRepository{}
		resp := doArticle(newTestRouter(fake), http.MethodPut, "/v1/article/"+articleID.String(), caseForm)

		require.Equal(t, http.StatusOK, resp.Code)
		require.JSONEq(t, `{"message":"Article updated"}`, resp.Body.String())
		require.Equal(t, 1, fake.updateCalls)
		require.Equal(t, identityTenant, fake.updateTenant)
		require.Equal(t, identityUser, fake.updateUser)
		require.Equal(t, articleID, fake.updateID)
		require.Equal(t, caseForm, fake.updateForm)
	})

	t.Run("zero rows update maps to 404", func(t *testing.T) {
		fake := &fakeRepository{updateErr: ErrNotFound}
		resp := doArticle(newTestRouter(fake), http.MethodPut, "/v1/article/"+articleID.String(), caseForm)

		require.Equal(t, http.StatusNotFound, resp.Code)
		require.JSONEq(t, `{"Message":"Article not found"}`, resp.Body.String())
	})

	t.Run("repository error", func(t *testing.T) {
		fake := &fakeRepository{updateErr: errors.New("boom")}
		resp := doArticle(newTestRouter(fake), http.MethodPut, "/v1/article/"+articleID.String(), caseForm)

		require.Equal(t, http.StatusNotAcceptable, resp.Code)
		require.JSONEq(t, `{"Message":"Article could not be updated"}`, resp.Body.String())
	})

	t.Run("binding error", func(t *testing.T) {
		fake := &fakeRepository{}
		resp := doArticle(newTestRouter(fake), http.MethodPut, "/v1/article/"+articleID.String(), map[string]string{"title": "ab", "content": "x"})

		require.Equal(t, http.StatusNotAcceptable, resp.Code)
		require.JSONEq(t, `{"message":"Title should be between 3 to 100 characters"}`, resp.Body.String())
	})

	t.Run("invalid parameter", func(t *testing.T) {
		fake := &fakeRepository{}
		resp := doArticle(newTestRouter(fake), http.MethodPut, "/v1/article/invalid", caseForm)

		require.Equal(t, http.StatusNotFound, resp.Code)
		require.JSONEq(t, `{"Message":"Invalid parameter"}`, resp.Body.String())
		require.Zero(t, fake.updateCalls)
	})
}

func TestDeleteArticle(t *testing.T) {
	t.Run("deleted", func(t *testing.T) {
		fake := &fakeRepository{}
		resp := doArticle(newTestRouter(fake), http.MethodDelete, "/v1/article/"+articleID.String(), nil)

		require.Equal(t, http.StatusOK, resp.Code)
		require.JSONEq(t, `{"message":"Article deleted"}`, resp.Body.String())
		require.Equal(t, 1, fake.deleteCalls)
		require.Equal(t, identityTenant, fake.deleteTenant)
		require.Equal(t, identityUser, fake.deleteUser)
		require.Equal(t, articleID, fake.deleteID)
	})

	t.Run("zero rows delete maps to 404", func(t *testing.T) {
		fake := &fakeRepository{deleteErr: ErrNotFound}
		resp := doArticle(newTestRouter(fake), http.MethodDelete, "/v1/article/"+articleID.String(), nil)

		require.Equal(t, http.StatusNotFound, resp.Code)
		require.JSONEq(t, `{"Message":"Article not found"}`, resp.Body.String())
	})

	t.Run("repository error", func(t *testing.T) {
		fake := &fakeRepository{deleteErr: errors.New("boom")}
		resp := doArticle(newTestRouter(fake), http.MethodDelete, "/v1/article/"+articleID.String(), nil)

		require.Equal(t, http.StatusNotAcceptable, resp.Code)
		require.JSONEq(t, `{"Message":"Article could not be deleted"}`, resp.Body.String())
	})

	t.Run("invalid parameter", func(t *testing.T) {
		fake := &fakeRepository{}
		resp := doArticle(newTestRouter(fake), http.MethodDelete, "/v1/article/invalid", nil)

		require.Equal(t, http.StatusNotFound, resp.Code)
		require.JSONEq(t, `{"Message":"Invalid parameter"}`, resp.Body.String())
		require.Zero(t, fake.deleteCalls)
	})
}
