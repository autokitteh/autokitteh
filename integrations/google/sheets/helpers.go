package sheets

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func a1RangeToGridRange(svc *sheets.Service, spreadsheetID, a1Range string) (*sheets.GridRange, error) {
	sheetName, from, to, err := a1RangeToSubstrings(a1Range)
	if err != nil {
		return nil, err
	}
	id, err := sheetID(svc, spreadsheetID, sheetName)
	if err != nil {
		return nil, err
	}
	gr := &sheets.GridRange{SheetId: id}

	col, row, err := cellToIndexes(from)
	if err != nil {
		return nil, err
	}
	if col >= 0 {
		gr.StartColumnIndex = col
	}
	if row >= 0 {
		gr.StartRowIndex = row
	}

	col, row, err = cellToIndexes(to)
	if err != nil {
		return nil, err
	}
	if col >= 0 {
		gr.EndColumnIndex = col + 1
	}
	if row >= 0 {
		gr.EndRowIndex = row + 1
	}

	return gr, nil
}

// https://developers.google.com/sheets/api/guides/concepts#expandable-1
func a1RangeToSubstrings(a1Range string) (string, string, string, error) {
	sheet := `['"]?(.+?)['"]?`
	// Accept lower-case or upper-case columns, but not mixes, so "Sheet1"
	// won't be interpreted as a cell.
	cells := `(([A-Z]*|[a-z]*)[0-9]*)(:(([A-Z]*|[a-z]*)[0-9]*))?`
	re := regexp.MustCompile(fmt.Sprintf(`^(%s!)?%s$`, sheet, cells))
	m := re.FindStringSubmatch(a1Range)

	if m == nil {
		re = regexp.MustCompile(fmt.Sprintf(`^%s$`, sheet))
		m = re.FindStringSubmatch(a1Range)
		if m == nil {
			return "", "", "", fmt.Errorf("invalid A1 range: %s", a1Range)
		}
		return m[1], "", "", nil
	}
	return m[2], m[3], m[6], nil
}

func sheetID(svc *sheets.Service, spreadsheetID, sheetName string) (int64, error) {
	ss, err := svc.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return 0, err
	}
	if len(ss.Sheets) == 0 {
		return 0, errors.New("no sheets found")
	}
	if sheetName == "" {
		return ss.Sheets[0].Properties.SheetId, nil
	}
	for _, s := range ss.Sheets {
		if s.Properties.Title == sheetName {
			return s.Properties.SheetId, nil
		}
	}
	return 0, fmt.Errorf("sheet %q not found in https://docs.google.com/spreadsheets/d/%s", sheetName, spreadsheetID)
}

func cellToIndexes(cell string) (int64, int64, error) {
	re := regexp.MustCompile(`([A-Za-z]*)([0-9]*)`)
	m := re.FindStringSubmatch(cell)
	if m == nil {
		return 0, 0, fmt.Errorf("invalid cell in A1 notation: %q", cell)
	}
	colStr, rowStr := m[1], m[2]
	col := lettersToBase26(strings.ToUpper(colStr))
	row, err := strconv.ParseInt(rowStr, 10, 64)
	if err != nil {
		row = 0 // rowStr == ""
	}
	return col - 1, row - 1, nil
}

func lettersToBase26(s string) int64 {
	if s == "" {
		return 0
	}
	result := int64(0)
	exp := float64(0)

	for i := len(s) - 1; i >= 0; i-- {
		c := int64(s[i] - 'A' + 1)
		result += c * int64(math.Pow(26, exp))
		exp++
	}
	return result
}

func indexToBase26(i int) string {
	i++
	result := ""
	for i > 0 {
		i--
		remainder := i % 26
		result = string(byte('A'+remainder)) + result
		i /= 26
	}
	return result
}

func hexToRGB(color int) (float64, float64, float64, float64, error) {
	if color < 0 || color > 0xFFFFFFFF {
		return 0, 0, 0, 0, fmt.Errorf("invalid RGBA value %d, should be between 0 and 0xFFFFFF (without alpha) or 0xFFFFFFFF (with alpha)", color)
	}
	if color <= 0xFFFFFF {
		color += (0xFF << 24) // Add default alpha if missing.
	}
	alpha := float64(color>>24) / 255
	red := float64((color>>16)&0xFF) / 255
	green := float64((color>>8)&0xFF) / 255
	blue := float64(color&0xFF) / 255
	return alpha, red, green, blue, nil
}
