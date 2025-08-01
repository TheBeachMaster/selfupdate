package service

import (
	"net/http"
)

type ServiceHandler interface {
	UpdateAppHandler() http.HandlerFunc
	CheckAppVersionHandler() http.HandlerFunc
}
