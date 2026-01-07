package inertia_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inertia "github.com/joetifa2003/inertigo"
	"github.com/joetifa2003/inertigo/vite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScroll_CreatesScrollProp(t *testing.T) {
	data := map[string]any{
		"data": []map[string]any{
			{"id": 1, "name": "First"},
			{"id": 2, "name": "Second"},
		},
	}
	config := inertia.ScrollConfig{
		PageName:     "page",
		CurrentPage:  1,
		NextPage:     2,
		PreviousPage: nil,
	}

	scrollProp := inertia.Scroll(data, config, "users.data")

	assert.Equal(t, data, scrollProp.Data)
	assert.Equal(t, config, scrollProp.Config)
	assert.Equal(t, "users.data", scrollProp.MergePath)
}

func TestRender_WithScrollProp_PopulatesScrollProps(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	data := map[string]any{
		"data": []map[string]any{
			{"id": 1, "title": "Post 1"},
			{"id": 2, "title": "Post 2"},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/posts?page=1", nil)
	req.Header.Set("X-Inertia", "true")

	rec := httptest.NewRecorder()

	err = i.Render(rec, req, "Posts/Index", inertia.Props{
		"posts": inertia.Scroll(data, inertia.ScrollConfig{
			PageName:     "page",
			CurrentPage:  1,
			NextPage:     2,
			PreviousPage: nil,
		}, "posts.data"),
	})
	require.NoError(t, err)

	var pageObject inertia.PageObject
	err = json.NewDecoder(rec.Body).Decode(&pageObject)
	require.NoError(t, err)

	// Verify scrollProps is populated
	assert.NotNil(t, pageObject.ScrollProps)
	postsScrollConfig, ok := pageObject.ScrollProps["posts"].(map[string]any)
	require.True(t, ok, "scrollProps should contain 'posts' key")
	assert.Equal(t, "page", postsScrollConfig["pageName"])
	assert.Equal(t, float64(1), postsScrollConfig["currentPage"]) // JSON unmarshals to float64
	assert.Equal(t, float64(2), postsScrollConfig["nextPage"])
	assert.Nil(t, postsScrollConfig["previousPage"])

	// Verify mergeProps is populated with append (default)
	assert.Contains(t, pageObject.MergeProps, "posts.data")

	// Verify the actual data is in props
	assert.NotNil(t, pageObject.Props["posts"])
}

func TestRender_WithScrollProp_PrependMergeIntent(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	data := map[string]any{
		"data": []map[string]any{
			{"id": 3, "message": "New message"},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/messages?page=2", nil)
	req.Header.Set("X-Inertia", "true")
	req.Header.Set("X-Inertia-Infinite-Scroll-Merge-Intent", "prepend")

	rec := httptest.NewRecorder()

	err = i.Render(rec, req, "Messages/Index", inertia.Props{
		"messages": inertia.Scroll(data, inertia.ScrollConfig{
			PageName:     "page",
			CurrentPage:  2,
			NextPage:     3,
			PreviousPage: 1,
		}, "messages.data"),
	})
	require.NoError(t, err)

	var pageObject inertia.PageObject
	err = json.NewDecoder(rec.Body).Decode(&pageObject)
	require.NoError(t, err)

	// Verify prependProps is populated instead of mergeProps
	assert.Contains(t, pageObject.PrependProps, "messages.data")
	assert.NotContains(t, pageObject.MergeProps, "messages.data")
}

func TestRender_WithMultipleScrollProps(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/feed", nil)
	req.Header.Set("X-Inertia", "true")

	rec := httptest.NewRecorder()

	err = i.Render(rec, req, "Feed/Index", inertia.Props{
		"posts": inertia.Scroll(
			map[string]any{"data": []any{}},
			inertia.ScrollConfig{PageName: "posts_page", CurrentPage: 1, NextPage: 2},
			"posts.data",
		),
		"comments": inertia.Scroll(
			map[string]any{"data": []any{}},
			inertia.ScrollConfig{PageName: "comments_page", CurrentPage: 1, NextPage: 2},
			"comments.data",
		),
	})
	require.NoError(t, err)

	var pageObject inertia.PageObject
	err = json.NewDecoder(rec.Body).Decode(&pageObject)
	require.NoError(t, err)

	// Verify both scroll props are present
	assert.NotNil(t, pageObject.ScrollProps["posts"])
	assert.NotNil(t, pageObject.ScrollProps["comments"])

	// Verify both merge paths are present
	assert.Contains(t, pageObject.MergeProps, "posts.data")
	assert.Contains(t, pageObject.MergeProps, "comments.data")
}

func TestRender_WithScrollPropAndRegularProps(t *testing.T) {
	bundler, err := vite.New(nil, vite.WithDevMode(true))
	require.NoError(t, err)

	i, err := inertia.New(bundler)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.Header.Set("X-Inertia", "true")

	rec := httptest.NewRecorder()

	err = i.Render(rec, req, "Posts/Index", inertia.Props{
		"posts": inertia.Scroll(
			map[string]any{"data": []any{}},
			inertia.ScrollConfig{PageName: "page", CurrentPage: 1, NextPage: 2},
			"posts.data",
		),
		"filters": map[string]string{"category": "tech"},
		"user":    map[string]any{"name": "John"},
	})
	require.NoError(t, err)

	var pageObject inertia.PageObject
	err = json.NewDecoder(rec.Body).Decode(&pageObject)
	require.NoError(t, err)

	// Verify scroll prop
	assert.NotNil(t, pageObject.ScrollProps["posts"])
	assert.Contains(t, pageObject.MergeProps, "posts.data")

	// Verify regular props are also present
	assert.NotNil(t, pageObject.Props["filters"])
	assert.NotNil(t, pageObject.Props["user"])
	assert.NotNil(t, pageObject.Props["posts"])
}
