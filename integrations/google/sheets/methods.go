package sheets

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/api/sheets/v4"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://developers.google.com/sheets/api/guides/concepts#expandable-1
func (a api) a1Range(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var sheetName, from, to string
	err := sdkmodule.UnpackArgs(args, kwargs,
		"sheet_name?", &sheetName,
		"from?", &from,
		"to?", &to,
	)
	if err != nil {
		return nil, err
	}
	if sheetName == "" && from == "" && to == "" {
		return nil, errors.New("no input")
	}
	if from == "" && to != "" {
		return nil, errors.New("to without from")
	}

	// Return the response.
	r := ""
	if sheetName != "" {
		// Single quotes are required for sheet names with spaces, special
		// characters, or an alphanumeric combination. Optional otherwise.
		r = fmt.Sprintf("'%s'", sheetName)
		if from == "" && to == "" {
			// Example: "Sheet1" - all the cells in "Sheet1".
			return sdkvalues.Wrap(r)
		}
		r += "!" // Possible to specify to/from cells without a sheet name.
	}
	r += from
	if to != "" {
		r = fmt.Sprintf("%s:%s", r, to)
	}
	return sdkvalues.Wrap(r)
}

// Read a single cell.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get
func (a api) readCell(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var spreadsheetID, sheetName, valueRenderOption string
	var rowIndex, colIndex int
	err := sdkmodule.UnpackArgs(args, kwargs,
		"spreadsheet_id", &spreadsheetID,
		"sheet_name?", &sheetName,
		"row_index", &rowIndex,
		"col_index", &colIndex,
		"value_render_option?", &valueRenderOption,
	)
	if err != nil {
		return nil, err
	}
	if rowIndex < 0 {
		return nil, fmt.Errorf("invalid row index: %d < 0", rowIndex)
	}
	if colIndex < 0 {
		return nil, fmt.Errorf("invalid column index: %d < 0", colIndex)
	}

	// Read a single-cell range of cells.
	kwargs = map[string]sdktypes.Value{
		"sheet_name": sdktypes.NewStringValue(sheetName),
		"from":       sdktypes.NewStringValue(indexToBase26(colIndex) + strconv.Itoa(rowIndex+1)),
	}
	singleCellRange, err := a.a1Range(ctx, []sdktypes.Value{}, kwargs)
	if err != nil {
		return nil, err
	}
	kwargs = map[string]sdktypes.Value{
		"spreadsheet_id":      sdktypes.NewStringValue(spreadsheetID),
		"a1_range":            singleCellRange,
		"value_render_option": sdktypes.NewStringValue(valueRenderOption),
	}
	v, err := a.readRange(ctx, []sdktypes.Value{}, kwargs)
	if err != nil {
		return nil, err
	}
	v = sdktypes.GetListValue(v)[0]
	v = sdktypes.GetListValue(v)[0]
	return v, nil
}

// Read a range of cells.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get
func (a api) readRange(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var spreadsheetID, a1Range, valueRenderOption string
	err := sdkmodule.UnpackArgs(args, kwargs,
		"spreadsheet_id", &spreadsheetID,
		"a1_range", &a1Range,
		"value_render_option?", &valueRenderOption,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.sheetsClient(ctx)
	if err != nil {
		return nil, err
	}
	call := client.Spreadsheets.Values.Get(spreadsheetID, a1Range)
	if valueRenderOption != "" {
		call.ValueRenderOption(valueRenderOption)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, err
	}

	// Normalize and return the response.
	maxRowLen := 0
	for _, row := range resp.Values {
		if len(row) > maxRowLen {
			maxRowLen = len(row)
		}
	}
	for i, row := range resp.Values {
		for j := len(row); j < maxRowLen; j++ {
			resp.Values[i] = append(resp.Values[i], "")
		}
	}
	return sdkvalues.Wrap(resp.Values)
}

// Write a single cell.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/update
func (a api) writeCell(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var spreadsheetID, sheetName string
	var rowIndex, colIndex int
	var value any
	err := sdkmodule.UnpackArgs(args, kwargs,
		"spreadsheet_id", &spreadsheetID,
		"sheet_name?", &sheetName,
		"row_index", &rowIndex,
		"col_index", &colIndex,
		"value", &value,
	)
	if err != nil {
		return nil, err
	}
	if rowIndex < 0 {
		return nil, fmt.Errorf("invalid row index: %d < 0", rowIndex)
	}
	if colIndex < 0 {
		return nil, fmt.Errorf("invalid column index: %d < 0", colIndex)
	}

	// Write a single-cell range of cells.
	kwargs = map[string]sdktypes.Value{
		"sheet_name": sdktypes.NewStringValue(sheetName),
		"from":       sdktypes.NewStringValue(indexToBase26(colIndex) + strconv.Itoa(rowIndex+1)),
	}
	singleCellRange, err := a.a1Range(ctx, []sdktypes.Value{}, kwargs)
	if err != nil {
		return nil, err
	}
	data, err := sdkvalues.Wrap([][]any{{value}})
	if err != nil {
		return nil, err
	}
	kwargs = map[string]sdktypes.Value{
		"spreadsheet_id": sdktypes.NewStringValue(spreadsheetID),
		"a1_range":       singleCellRange,
		"data":           data,
	}
	return a.writeRange(ctx, []sdktypes.Value{}, kwargs)
}

// Write a range of cells.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/update
func (a api) writeRange(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var spreadsheetID, a1Range string
	var wrappedData sdktypes.Value
	err := sdkmodule.UnpackArgs(args, kwargs,
		"spreadsheet_id", &spreadsheetID,
		"a1_range", &a1Range,
		"data", &wrappedData,
	)
	if err != nil {
		return nil, err
	}

	// Unwrap input data to a 2D matrix that is acceptable by the API
	// ("sdkmodule.UnpackArgs()" directly into "[][]any" doesn't work).
	if !sdktypes.IsListValue(wrappedData) {
		return nil, errors.New("invalid data")
	}
	rows := sdktypes.GetListValue(wrappedData)
	data, err := kittehs.TransformError(rows, func(row sdktypes.Value) ([]any, error) {
		if !sdktypes.IsListValue(row) {
			return nil, errors.New("invalid data")
		}
		cols := sdktypes.GetListValue(row)
		return kittehs.TransformError(cols, func(col sdktypes.Value) (any, error) {
			return sdkvalues.Unwrap(col)
		})
	})
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.sheetsClient(ctx)
	if err != nil {
		return nil, err
	}
	call := client.Spreadsheets.Values.Update(spreadsheetID, a1Range, &sheets.ValueRange{
		Range:  a1Range,
		Values: data,
	})
	resp, err := call.ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Set the background color in a range of cells.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/batchUpdate
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#CellData
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#CellFormat
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/other#ColorStyle
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/other#Color
func (a api) setBackgroundColor(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var spreadsheetID, a1Range string
	var color int
	err := sdkmodule.UnpackArgs(args, kwargs,
		"spreadsheet_id", &spreadsheetID,
		"a1_range", &a1Range,
		"color", &color,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.sheetsClient(ctx)
	if err != nil {
		return nil, err
	}
	gr, err := a1RangeToGridRange(client, spreadsheetID, a1Range)
	if err != nil {
		return nil, err
	}
	alpha, red, green, blue, err := hexToRGB(color)
	if err != nil {
		return nil, err
	}
	resp, err := client.Spreadsheets.BatchUpdate(spreadsheetID,
		&sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{{
				RepeatCell: &sheets.RepeatCellRequest{
					Cell: &sheets.CellData{
						UserEnteredFormat: &sheets.CellFormat{
							BackgroundColorStyle: &sheets.ColorStyle{
								RgbColor: &sheets.Color{
									Alpha: alpha,
									Red:   red,
									Green: green,
									Blue:  blue,
								},
							},
						},
					},
					Fields: "userEnteredFormat.backgroundColorStyle",
					Range:  gr,
				},
			}},
		}).Do()
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Set the text format in a range of cells.
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/batchUpdate
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#CellData
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/cells#CellFormat
// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets/other#TextFormat
func (a api) setTextFormat(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var spreadsheetID, a1Range string
	var color int
	var bold, italic, strikethrough, underline bool
	err := sdkmodule.UnpackArgs(args, kwargs,
		"spreadsheet_id", &spreadsheetID,
		"a1_range", &a1Range,
		"color?", &color,
		"bold?", &bold,
		"italic?", &italic,
		"strikethrough?", &strikethrough,
		"underline?", underline,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.sheetsClient(ctx)
	if err != nil {
		return nil, err
	}
	gr, err := a1RangeToGridRange(client, spreadsheetID, a1Range)
	if err != nil {
		return nil, err
	}
	alpha, red, green, blue, err := hexToRGB(color)
	if err != nil {
		return nil, err
	}
	resp, err := client.Spreadsheets.BatchUpdate(spreadsheetID,
		&sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{{
				RepeatCell: &sheets.RepeatCellRequest{
					Cell: &sheets.CellData{
						UserEnteredFormat: &sheets.CellFormat{
							TextFormat: &sheets.TextFormat{
								ForegroundColorStyle: &sheets.ColorStyle{
									RgbColor: &sheets.Color{
										Alpha: alpha,
										Red:   red,
										Green: green,
										Blue:  blue,
									},
								},
								Bold:          bold,
								Italic:        italic,
								Strikethrough: strikethrough,
								Underline:     underline,
							},
						},
					},
					Fields: "userEnteredFormat.textFormat",
					Range:  gr,
				},
			}},
		}).Do()
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}
