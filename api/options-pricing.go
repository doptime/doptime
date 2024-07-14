package api

type PricingByToken struct {
	PricePerRequestToken  float64
	PricePerResponseToken float64
}
type PricingByKB struct {
	PricePerRequestKB  float64
	PricePerResponseKB float64
}

// should allow pay to api VendorAccountEmail ˝
type VendorInfo struct {
	PricingByCall  float64
	PricingByToken *PricingByToken
	PricingByKB    *PricingByKB

	VendorAccountEmail string `json:"-"`
	VendorIMs          string
}
