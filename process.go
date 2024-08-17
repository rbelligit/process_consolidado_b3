package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/rbelligit/process_consolidado_b3/process_consolidado"
)

func main() {
	folder := flag.String("folder", "./", "Pasta onde estão os arquivos")
	files := flag.String("files", "", "Lista de arquivos separados por ,")
	result := flag.String("result", "result.csv", "Onde colocar o resultado em CSV")

	flag.Parse()

	filesList := strings.Split(*files, ",")
	filesPath := make([]string, 0, len(filesList))
	for _, file := range filesList {
		filesPath = append(filesPath, path.Join(*folder, file))
	}

	fmt.Fprintf(os.Stderr, "Files: [%+v]\n", filesPath)

	proc, err := process_consolidado.NewProcessConsolidado(filesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro=%s\n", err.Error())
		return
	}
	invest, err := proc.Process()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro Process: %s\n", err.Error())
		return
	}

	tipos, err := process_consolidado.NewTiposFiis("./FIIs_classificacao.xlsx")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading tipos: %s\n", err.Error())
		return
	}

	genOutput(invest, path.Join(*folder, *result), tipos)
}

func processFiis(fiis *process_consolidado.InvestimentsData, tipos *process_consolidado.TiposFiis) map[string]*process_consolidado.InvestimentsData {
	var ans map[string]*process_consolidado.InvestimentsData
	ans = make(map[string]*process_consolidado.InvestimentsData)
	for n, v := range *fiis {
		if vv, ok := tipos.Tipos[n]; ok {
			if inv, ok := ans[vv]; ok {
				(*inv)[n] = v
			} else {
				val := make(process_consolidado.InvestimentsData)
				ans[vv] = &val
				val[n] = v
			}
		} else {
			fmt.Fprintf(os.Stderr, "%s SEM TIPO!!!!\n", n)
		}
	}
	return ans
}

func calcTotal(inv *process_consolidado.Investments) process_consolidado.Reais {
	total := process_consolidado.Reais(0)
	for _, v := range inv.Acoes {
		total += v.PricePerItem * process_consolidado.Reais(v.Quantity)
	}
	for _, v := range inv.Fiis {
		total += v.PricePerItem * process_consolidado.Reais(v.Quantity)
	}
	return total
}

type PercentagemBase100 int // Valor de 0 até 10000 -> 100%

func calcPercentagem(valor int, total int) PercentagemBase100 {
	res := int64(valor) * int64(10000) / int64(total)
	return PercentagemBase100(res)
}

func (p PercentagemBase100) ToString() string {
	return fmt.Sprintf("%d.%02d", p/100, p%100)
}

func genOutput(invest *process_consolidado.Investments, fileName string, tipos *process_consolidado.TiposFiis) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("open file %w", err)
	}

	fmt.Fprintf(file, "Tipo, Código, Quantidade, Preço unitário, Preço total, Percentagem, percentagem do Tipo\n")

	total := calcTotal(invest)
	first := true
	totalAcoes := process_consolidado.Reais(0)
	for _, v := range invest.Acoes {
		if !first {
			fmt.Fprintf(file, "\n")
		}
		first = false
		valorAt := v.Quantity * int(v.PricePerItem)
		porAt := calcPercentagem(valorAt, int(total))
		totalAcoes += process_consolidado.Reais(valorAt)
		fmt.Fprintf(file, "ações,%s, %d, %f, %f, %s", v.Codigo, v.Quantity, float64(int(v.PricePerItem))/100.0, float64(valorAt)/100.0, porAt.ToString())
	}
	fmt.Fprintf(file, ", %s\n", calcPercentagem(int(totalAcoes), int(total)).ToString())
	fiisProc := processFiis(&invest.Fiis, tipos)
	for tipo, invs := range fiisProc {
		first = true
		totalTipo := process_consolidado.Reais(0)
		for _, inv := range *invs {
			if !first {
				fmt.Fprintf(file, "\n")
			}
			first = false
			valorAt := inv.Quantity * int(inv.PricePerItem)
			totalTipo += process_consolidado.Reais(valorAt)
			fmt.Fprintf(file, "%s, %s, %d, %f, %f, %s", tipo, inv.Codigo, inv.Quantity,
				float64(int(inv.PricePerItem))/100.0, float64(valorAt)/100.0,
				calcPercentagem(int(valorAt), int(total)).ToString())
		}
		fmt.Fprintf(file, ", %s\n", calcPercentagem(int(totalTipo), int(total)).ToString())
	}
	return nil
}
