package core

/*
 * File: pkg/core/routes.go
 *
 * Purpose: definition of the server routing rules
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import "github.com/gin-gonic/gin"

func (srv *tournabyteAPIService) addGlobalMiddleware() {
	srv.router.Use(gin.CustomRecovery(srv.recoverPanicAsFailure))
	srv.router.Use(srv.assignRequestIdentifier)
	srv.router.Use(srv.markRequestStartTimestamp)
}

func (srv *tournabyteAPIService) registerRoutes() {
	srv.addGlobalMiddleware()

	{
		// /v1/...
		v1 := srv.router.Group("v1")
		srv.addAuthGroup(v1)
	}
}

func (srv *tournabyteAPIService) addAuthGroup(parentGroup *gin.RouterGroup) {
	parentGroup.Group("auth")
}
