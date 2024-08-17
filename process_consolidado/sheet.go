package process_consolidado

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thedatashed/xlsxreader"
)

const _acoes = 1
const _fiis = 2

var _sheets = map[string]int{
	"Posição - Ações":  1,
	"Posição - Fundos": 2,
}

type sheet struct {
	sheetFile string
}

func newSheet(fileName string) *sheet {
	return &sheet{
		sheetFile: fileName,
	}
}

func (s *sheet) processSheet() (InvestimentsData /*acoes*/, InvestimentsData /*fiis*/, error) {
	xl, err := xlsxreader.OpenFile(s.sheetFile)
	if err != nil {
		return nil, nil, err
	}

	result := make(map[int]InvestimentsData)
	for _, name := range xl.Sheets {
		if tipo, ok := _sheets[name]; ok {
			fmt.Fprintf(os.Stderr, "Vai proc sheet %s\n", name)
			invests, err := s.processPage(xl, name)
			if err != nil {
				return nil, nil, fmt.Errorf("proc sheet %s - page %s: %w", s.sheetFile, name, err)
			}
			result[tipo] = invests
		}
	}
	return result[_acoes], result[_fiis], nil
}

func getCells(cells []xlsxreader.Cell) map[string]string {
	ans := make(map[string]string)
	for i := range cells {
		ans[strings.ToUpper(cells[i].Column)] = strings.Trim(cells[i].Value, " \t\r\n")
	}
	return ans
}

func AddInvestments(v1 *InvestmentData, v2 *InvestmentData) *InvestmentData {

	v1.Quantity += v2.Quantity
	return v1
}

func addToAns(m *InvestimentsData, inv *InvestmentData) {
	if v, ok := (*m)[inv.Codigo]; ok {
		(*m)[inv.Codigo] = AddInvestments(inv, v)
	} else {
		(*m)[inv.Codigo] = inv
	}
}

func (s *sheet) processPage(xl *xlsxreader.XlsxFileCloser, pageName string) (InvestimentsData, error) {
	ans := make(InvestimentsData)
	for row := range xl.ReadRows(pageName) {
		//fmt.Fprintf(os.Stderr, "read row: %d\n", row.Index)
		if row.Index > 1 {
			cels := getCells(row.Cells)
			if len(cels) < 5 {
				continue
			}
			if len(cels["A"]) < 4 {
				continue
			}
			//fmt.Fprintf(os.Stderr, "%d: Cells: %+v\n", row.Cells[0].Row, cels)
			quant, err := strconv.ParseInt(cels["I"], 10, 32)
			if err != nil {
				return nil, fmt.Errorf(" reading quantity row=%d - err: %s", row.Index, err.Error())
			}
			pricef, err := strconv.ParseFloat(cels["M"], 64)
			if err != nil {
				return nil, fmt.Errorf(" reading value row=%d - err: %s", row.Index, err.Error())
			}
			priceR := Reais(pricef * 100)

			dt := InvestmentData{
				Codigo:       cels["D"],
				Quantity:     int(quant),
				PricePerItem: priceR,
				Produto:      cels["A"],
				Cnpj:         cels["E"],
				AdmEscr:      cels["H"],
			}

			addToAns(&ans, &dt)
		}
	}
	return ans, nil
}
