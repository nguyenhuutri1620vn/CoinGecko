package models

import (
	"time"
)

type PriceHistory struct {
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	Time   int64   `json:"time"`
	Change float64 `json:"change"`
}

type Input struct {
	Symbol string    `json:"sympol_input" form:"sympol_input" query:"sympol_input"`
	Start  time.Time `json:"start_date" form:"start_date" query:"start_date"`
	End    time.Time `json:"end_date" form:"end_date" query:"end_date"`
	Period string    `json:"period" form:"period" query:"period"`
}

type ResponseData struct {
	Prices     [][]float64 `json:"prices"`
	MarKetCaps [][]float32 `json:"market_caps"`
	Volume     [][]float32 `json:"Volume"`
}

func (p *PriceHistory) CalculateChange(previous *PriceHistory) float64 {
	if previous == nil {
		return 0
	}
	return (p.Close - previous.Close) / previous.Close * 100
}
