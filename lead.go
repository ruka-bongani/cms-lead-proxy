package main

// LeadWrapper & nested types for the CMS payload
type LeadWrapper struct {
	Lead Lead `json:"lead"`
}

type Lead struct {
	DealerRef         string            `json:"dealerRef"`
	DealerFloor       string            `json:"dealerFloor"`
	DealerSalesPerson string            `json:"dealerSalesPerson"`
	Region            string            `json:"region"`
	Source            string            `json:"source"`
	TransactionID     string            `json:"transactionID"`
	ExtLeadRef        string            `json:"extLeadRef"`
	PromotionalCode   string            `json:"promotionalCode"`
	UtmParameters     string            `json:"utmParameters"`
	CountryCode       string            `json:"countryCode"`
	Contact           Contact           `json:"contact"`
	Seeks             Seeks             `json:"seeks"`
	Referrer          map[string]string `json:"referrer"`
	Options           map[string]string `json:"options"`
	TradeIns          []TradeIn         `json:"TradeIns"`
}

type Seeks struct {
	Used             string `json:"used"`
	Brand            string `json:"brand"`
	Model            string `json:"model"`
	MmCode           string `json:"mmCode"`
	ModelCode        string `json:"modelCode"`
	Kms              string `json:"kms"`
	Year             string `json:"year"`
	Colour           string `json:"colour"`
	StockNr          string `json:"stockNr"`
	Price            string `json:"price"`
	Deposit          string `json:"deposit"`
	TestDrive        string `json:"testDrive"`
	TradeIn          string `json:"tradeIn"`
	Finance          string `json:"finance"`
	Valuation        string `json:"valuation"`
	Registration     string `json:"registration"`
	Special          string `json:"special"`
	SpecialBannerURL string `json:"specialBannerURL"`
	ServiceHistory   string `json:"serviceHistory"`
	Comments         string `json:"comments"`
	Vin              string `json:"vin"`
}

type TradeIn struct {
	Make       string  `json:"Make"`
	Model      string  `json:"Model"`
	Variant    string  `json:"Variant"`
	Year       int     `json:"Year"`
	Mileage    int     `json:"Mileage"`
	MMCode     *int    `json:"MMCode"`
	IsFinanced bool    `json:"IsFinanced"`
	Price      float64 `json:"Price"`
}

type CMSResponse struct {
	Code          string `json:"code"`
	LeadReference string `json:"leadReference"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}
