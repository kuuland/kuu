package kuu

import "github.com/kuuland/kuu/route"

// RouteInfo represents a request route's specification which contains method and path and its handler.
type RouteInfo struct {
	Method         string
	Path           string
	Name           string
	HandlerFunc    HandlerFunc
	IgnorePrefix   bool
	Description    string
	Tags           []string
	RequestParams  route.RequestParams
	ResponseParams route.ResponseParams
}

// RoutesInfo defines a RouteInfo array.
type RoutesInfo []RouteInfo
