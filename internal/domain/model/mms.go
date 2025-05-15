package model

import "time"

// MMS representa uma média móvel simples para um par em um timestamp específico
type MMS struct {
	Pair      string    // Par de moedas (BRLBTC, BRLETH)
	Timestamp time.Time // Data da MMS
	MMS20     float64   // Média móvel simples de 20 dias
	MMS50     float64   // Média móvel simples de 50 dias
	MMS200    float64   // Média móvel simples de 200 dias
}

// Período válido para cálculo de MMS
const (
	Period20  = 20
	Period50  = 50
	Period200 = 200
)

// Validar se o período solicitado é válido
func IsValidPeriod(period int) bool {
	return period == Period20 || period == Period50 || period == Period200
}

// Validar se o par de moedas é válido
func IsValidPair(pair string) bool {
	return pair == "BRLBTC" || pair == "BRLETH"
}
