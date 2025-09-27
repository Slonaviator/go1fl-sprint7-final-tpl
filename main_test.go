package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "tula"
	minTotalNumCafe := min(len(cafeList[city]), 100) // минимальное значение из двух: факт. количество кафе и 100

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: minTotalNumCafe},
	}

	for _, r := range requests {
		//"count="+string(r.count)
		t.Run(fmt.Sprintf("count=%d", r.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?count=%d&city=%s", r.count, city), nil)

			handler.ServeHTTP(response, req)
			require.Equal(t, http.StatusOK, response.Code)
			body := strings.TrimSpace(response.Body.String())

			if r.want == 0 {
				assert.Empty(t, body)
				return
			}
			cafes := strings.Split(body, ",")
			assert.Len(t, cafes, r.want)
		})
	}

}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"

	requests := []struct {
		search string
		want   int
	}{
		{search: "фасоль", want: 0},
		{search: "кофе", want: 2},
		{search: "вилка", want: 1},
	}

	for _, r := range requests {
		t.Run(fmt.Sprintf("search=%s", r.search), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=%s&search=%s", city, r.search), nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)
			body := strings.TrimSpace(response.Body.String())
			if r.want == 0 {
				assert.Empty(t, body)
				return
			}
			cafes := strings.Split(body, ",")
			assert.Len(t, cafes, r.want)

			for _, cafe := range cafes {
				assert.Contains(t, strings.ToLower(cafe), strings.ToLower(r.search))
			}
		})
	}
}
