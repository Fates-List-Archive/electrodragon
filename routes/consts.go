package routes

import (
	"net/http"
	"wv2/types"

	jsoniter "github.com/json-iterator/go"
)

// A much faster json parser for all routes
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Defines error constants
const (
	notFoundPage  = "Not Found"
	internalError = "Something went wrong"
	invalidMethod = "Invalid method"
)

// Defines a function for a route
type RouteFunc func(w http.ResponseWriter, r *http.Request, opts types.RouteInfo)
