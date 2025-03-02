package models

import "gorm.io/gorm"

type RequestLog struct {
	gorm.Model
	Method     string `json:"method"`
	Endpoint   string `json:"endpoint"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	StatusCode int    `json:"status_code"`
}
