package httpx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestMapPgxError(t *testing.T) {
	pgErr := func(code string) *pgconn.PgError {
		return &pgconn.PgError{Code: code}
	}

	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"missing row", pgx.ErrNoRows, http.StatusNotFound},
		{"wrapped missing row", fmt.Errorf("get: %w", pgx.ErrNoRows), http.StatusNotFound},
		{"unique violation", pgErr("23505"), http.StatusConflict},
		{"foreign key violation", pgErr("23503"), http.StatusConflict},
		{"rls violation", pgErr("42501"), http.StatusInternalServerError},
		{"other pgx code", pgErr("22001"), http.StatusInternalServerError},
		{"wrapped pgx error", fmt.Errorf("list: %w", pgErr("23505")), http.StatusConflict},
		{"plain error", fmt.Errorf("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.status, MapPgxError(tc.err))
		})
	}
}

func TestParseUUIDParam(t *testing.T) {
	router := gin.New()
	var parsed uuid.UUID
	router.GET("/article/:id", func(c *gin.Context) {
		id, err := ParseUUIDParam(c, "id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
			return
		}
		parsed = id
		c.Status(http.StatusOK)
	})

	cases := []struct {
		path      string
		status    int
		wantParse uuid.UUID
	}{
		{"/article/" + testID(), http.StatusOK, uuid.MustParse(testID())},
		{"/article/invalid", http.StatusBadRequest, uuid.Nil},
		{"/article/00000000-0000-0000-0000-000000000000", http.StatusOK, uuid.Nil},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			parsed = uuid.Nil
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, tc.path, nil))
			require.Equal(t, tc.status, resp.Code)
			require.Equal(t, tc.wantParse, parsed)
		})
	}
}

func TestAbortMessage(t *testing.T) {
	router := gin.New()
	router.GET("/boom", func(c *gin.Context) {
		AbortMessage(c, http.StatusNotAcceptable, "Invalid login details")
	})

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/boom", nil))
	require.Equal(t, http.StatusNotAcceptable, resp.Code)
	require.JSONEq(t, `{"message":"Invalid login details"}`, resp.Body.String())
}

func TestJSONMessageDoesNotAbort(t *testing.T) {
	router := gin.New()
	router.GET("/ok", func(c *gin.Context) {
		JSONMessage(c, http.StatusOK, "Article updated")
	})

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/ok", nil))
	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, `{"message":"Article updated"}`, resp.Body.String())
}

// testID is a fixed UUID string shared by the table above.
func testID() string { return "e7b8f04d-3a2b-44d5-a0c3-8e11bd2b9a01" }
