package config

import (
	"github.com/go-resty/resty/v2"
)

var (
	RestClient *resty.Client
)

func init() {
	RestClient = resty.New().
		SetHeader("User-Agent", "Axon CI Server").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(1000).
		SetRetryMaxWaitTime(5000).
		SetTimeout(30 * 1000) // 30 seconds timeout
}
