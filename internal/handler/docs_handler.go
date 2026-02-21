package handler

import (
	"net/http"

	_ "github.com/HyGames-fun/hytale-monitoring-backend/docs"
	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/swaggo/swag"
)

var scalarHTML string

func init() {
	doc, err := swag.ReadDoc()
	if err != nil {
		panic(err)
	}

	scalarHTML, err = scalar.ApiReferenceHTML(&scalar.Options{
		SpecContent:  doc,
		Theme:        scalar.ThemePurple,
		Layout:       scalar.LayoutModern,
		DarkMode:     true,
		ShowSidebar:  true,
		SearchHotKey: "k",
		CustomOptions: scalar.CustomOptions{
			PageTitle: "Hytale Monitoring API Reference",
		},
	})
	if err != nil {
		panic(err)
	}
}

func ScalarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(scalarHTML))
}

func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	doc, _ := swag.ReadDoc()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(doc))
}
