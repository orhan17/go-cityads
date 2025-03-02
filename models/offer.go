package models

type Offer struct {
	ExternalID   int     `gorm:"primaryKey" json:"external_id"`
	Name         string  `json:"name"`
	Currency     string  `json:"currency"`
	ApprovalTime int     `json:"approval_time"`
	SiteURL      string  `json:"site_url"`
	Logo         string  `json:"logo"`
	GeoCode      string  `json:"geo_code"`
	GeoName      string  `json:"geo_name"`
	Rating       float64 `json:"rating"`
}
