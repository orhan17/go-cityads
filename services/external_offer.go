package services

type ExternalOffer struct {
	ExternalID    string `json:"id"`
	Name          string `json:"name"`
	OfferCurrency struct {
		Name string `json:"name"`
	} `json:"offer_currency"`
	ApprovalTime string `json:"approval_time"`
	PaymentTime  string `json:"payment_time"`
	SiteURL      string `json:"site_url"`
	Logo         string `json:"logo"`
	Geo          []struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"geo"`
	Stat struct {
		ECPL string `json:"ecpl"`
	} `json:"stat"`
}

type APIResponse struct {
	Offers []ExternalOffer `json:"offers"`
}
