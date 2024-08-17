package process_consolidado

import (
	"fmt"
	"os"
)

type ProcessConsolidado struct {
	sheetsNames []string
}

type Reais int

type InvestmentData struct {
	Codigo       string
	Quantity     int
	PricePerItem Reais
	Produto      string
	Cnpj         string
	AdmEscr      string
}

type InvestimentsData map[string]*InvestmentData

type Investments struct {
	Fiis  InvestimentsData
	Acoes InvestimentsData
}

func NewProcessConsolidado(sheets []string) (*ProcessConsolidado, error) {
	proc := &ProcessConsolidado{
		sheetsNames: sheets,
	}

	return proc, nil
}

func (pc *ProcessConsolidado) Process() (*Investments, error) {
	var invCons Investments
	invCons.Acoes = make(InvestimentsData)
	invCons.Fiis = make(InvestimentsData)
	for _, planilhas := range pc.sheetsNames {
		fmt.Fprintf(os.Stderr, "### Iniciando processamento da planilha %s\n", planilhas)
		s := newSheet(planilhas)
		acoes, fiis, err := s.processSheet()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro processSheet: %s\n", err.Error())
			return nil, err
		}

		invCons.Accumulate(&acoes, &fiis)
	}
	return &invCons, nil
}

func (i *Investments) Accumulate(acoes *InvestimentsData, fiis *InvestimentsData) {
	fmt.Fprintf(os.Stderr, "acoes quant=%d - fiis quant=%d\n", len(*acoes), len(*fiis))

	for nome, invest := range *acoes {
		if v, ok := i.Acoes[nome]; ok {
			v.Quantity += invest.Quantity
		} else {
			i.Acoes[nome] = invest
		}
	}
	for nome, invest := range *fiis {
		if v, ok := i.Fiis[nome]; ok {
			v.Quantity += invest.Quantity
		} else {
			i.Fiis[nome] = invest
		}
	}
}
