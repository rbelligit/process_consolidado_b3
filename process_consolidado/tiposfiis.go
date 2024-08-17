package process_consolidado

import (
	"fmt"
	"os"

	"github.com/thedatashed/xlsxreader"
)

type TiposFiis struct {
	Tipos map[string]string
}

func NewTiposFiis(file string) (*TiposFiis, error) {
	tipo := &TiposFiis{
		Tipos: make(map[string]string),
	}
	reader, err := xlsxreader.OpenFile(file)

	if err != nil {
		return nil, fmt.Errorf("abrindo xls: %w", err)
	}
	err = tipo.readTipos(reader)
	return tipo, err
}

func (t *TiposFiis) readTipos(reader *xlsxreader.XlsxFileCloser) error {
	sh := reader.Sheets[0]
	fmt.Fprintf(os.Stderr, "Lendo sheet: %s\n", sh)

	for rows := range reader.ReadRows(sh) {
		if rows.Index > 1 {
			cels := getCells(rows.Cells)
			t.Tipos[cels["A"]] = cels["C"]
		}
	}
	return nil
}
