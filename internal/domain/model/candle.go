package model

import "time"

// Candle representa um candle de mercado em um determinado período
type Candle struct {
	Pair      string    // Par de moedas (BRLBTC, BRLETH)
	Timestamp time.Time // Data/hora do candle
	Open      float64   // Preço de abertura
	High      float64   // Preço máximo
	Low       float64   // Preço mínimo
	Close     float64   // Preço de fechamento
	Volume    float64   // Volume negociado
}
