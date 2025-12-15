package model

// Geolocation represents the geographic and network information
// associated with an IP address. This is the normalised result type
// that all checkers map their responses to.
type Geolocation struct {
	// IP is the queried address
	IP IPAddress `json:"ip"`

	// Geographic information
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`

	// Network information
	ISP string `json:"isp"`
	Org string `json:"org"`
	ASN string `json:"asn"`
}

// HasLocation reports whether the geolocation has valid coordinates.
func (g Geolocation) HasLocation() bool {
	return g.Latitude != 0 || g.Longitude != 0
}

// HasNetworkInfo reports whether the geolocation has any network information.
func (g Geolocation) HasNetworkInfo() bool {
	return g.ISP != "" || g.Org != "" || g.ASN != ""
}

// IsEmpty reports whether all fields (except IP) are at their zero values.
func (g Geolocation) IsEmpty() bool {
	return g.Country == "" &&
		g.CountryCode == "" &&
		g.Region == "" &&
		g.City == "" &&
		g.Latitude == 0 &&
		g.Longitude == 0 &&
		g.ISP == "" &&
		g.Org == "" &&
		g.ASN == ""
}
