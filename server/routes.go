package main

import (
	"github.com/MerlinDMC/dsapid/server/handler"
	"github.com/codegangsta/martini"
)

func registerRoutes(router martini.Router) {
	// common
	router.Get("/ping", handler.CommonPing)
	router.Get("/status", handler.CommonStatus)

	// dsapi
	router.Get("/datasets", handler.DsapiList)
	router.Get("/datasets/:id", handler.DsapiDetail)
	router.Get("/datasets/:id/:path", handler.DsapiFile)

	// imgapi
	router.Get("/images", handler.ImgapiList)
	router.Get("/images/:id", handler.ImgapiDetail)
	router.Get("/images/:id/file", handler.ImgapiFile)
	router.Get("/images/:id/file:file_idx", handler.ImgapiFile)

	// public api
	router.Get("/api/datasets", handler.ApiDatasetsList)
	router.Get("/api/export/:id", handler.ApiDatasetExport)
}
