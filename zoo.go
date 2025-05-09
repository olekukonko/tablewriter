package tablewriter

import (
	"database/sql"
	"fmt"
	"github.com/olekukonko/errors"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"reflect"
	"strings"
)

// rawCellsToStrings converts a row to its raw string representation using specified cell config for filters.
// 'rowInput' can be []string, []any, or a custom type if t.stringer is set.
func (t *Table) rawCellsToStrings(rowInput interface{}, cellCfg tw.CellConfig) ([]string, error) {
	t.logger.Debug("rawCellsToStrings: Converting row: %v (type: %T) using filters from specified CellConfig", rowInput, rowInput)
	var cells []string
	var conversionError error

	switch v := rowInput.(type) {
	case []string:
		cells = v
		t.logger.Debug("rawCellsToStrings: Input is []string.")
	case []any:
		cells = make([]string, len(v))
		for i, element := range v {
			cells[i] = t.convertToString(element)
		}
		t.logger.Debug("rawCellsToStrings: Input is []any, processed elements.")
	default:
		// var ok bool
		if t.stringer != nil {
			cells, conversionError = t.callStringer(rowInput)
			if conversionError != nil {
				t.logger.Debug("rawCellsToStrings: Stringer error: %v", conversionError)
				return nil, conversionError
			}
			t.logger.Debug("rawCellsToStrings: Input (custom type) converted to: %v", cells)
		} else if stringer, ok := rowInput.(fmt.Stringer); ok {
			cells = []string{stringer.String()}
			t.logger.Debug("rawCellsToStrings: Input is fmt.Stringer, used String() for single cell")
		} else {
			conversionError = errors.Newf("cannot convert row type %T to []string; not []string, []any, no t.stringer, not fmt.Stringer", rowInput)
			t.logger.Debug("rawCellsToStrings: Conversion error: %v", conversionError)
			return nil, conversionError
		}
	}

	if cellCfg.Filter.Global != nil {
		cells = cellCfg.Filter.Global(cells)
		t.logger.Debug("rawCellsToStrings: Applied global filter. Result  Result: %v", cells)
	}

	if len(cellCfg.Filter.PerColumn) > 0 {
		originalCellsForLog := append([]string(nil), cells...)
		changedByFilter := false
		for i := 0; i < len(cells) && i < len(cellCfg.Filter.PerColumn); i++ {
			if filter := cellCfg.Filter.PerColumn[i]; filter != nil {
				newCell := filter(cells[i])
				if newCell != cells[i] {
					changedByFilter = true
				}
				cells[i] = newCell
			}
		}
		if changedByFilter {
			t.logger.Debug("rawCellsToStrings: Applied per-column filters. Original: %v, Result: %v", originalCellsForLog, cells)
		}
	}
	return cells, nil
}

// toStringLines converts raw cells to formatted lines for table output
func (t *Table) toStringLines(row interface{}, config tw.CellConfig) ([][]string, error) {
	cells, err := t.rawCellsToStrings(row, config)
	if err != nil {
		return nil, err
	}
	return t.prepareContent(cells, config), nil
}

// processVariadicElements handles the common logic for processing variadic arguments
// that could be either individual elements or a slice of elements
func (t *Table) processVariadicElements(elements []any) []any {
	// Check if the input looks like a single slice was passed
	if len(elements) == 1 {
		firstArg := elements[0]
		// Try to assert to common slice types users might pass
		switch v := firstArg.(type) {
		case []string:
			t.logger.Debug("Detected single []string argument. Unpacking it.")
			result := make([]any, len(v))
			for i, val := range v {
				result[i] = val
			}
			return result
		case []interface{}:
			t.logger.Debug("Detected single []interface{} argument. Unpacking it.")
			result := make([]any, len(v))
			for i, val := range v {
				result[i] = val
			}
			return result
		}
	}

	// Either multiple arguments were passed, or a single non-slice argument
	t.logger.Debug("Input has multiple elements, is empty, or is a single non-slice element. Using variadic elements as is.")
	return elements
}

// prepareTableSection prepares either headers or footers for the table
func (t *Table) prepareTableSection(elements []any, config tw.CellConfig, sectionName string) [][]string {
	actualCellsToProcess := t.processVariadicElements(elements)
	t.logger.Debug("%s(): Effective cells to process: %v", sectionName, actualCellsToProcess)

	stringsResult, err := t.rawCellsToStrings(actualCellsToProcess, config)
	if err != nil {
		t.logger.Error("%s(): Failed to convert elements to strings: %v", sectionName, err)
		stringsResult = []string{}
	}

	prepared := t.prepareContent(stringsResult, config)
	numColsBatch := t.maxColumns()

	if len(prepared) > 0 {
		for i := range prepared {
			if len(prepared[i]) < numColsBatch {
				t.logger.Debug("Padding %s line %d from %d to %d columns", sectionName, i, len(prepared[i]), numColsBatch)
				paddedLine := make([]string, numColsBatch)
				copy(paddedLine, prepared[i])
				for j := len(prepared[i]); j < numColsBatch; j++ {
					paddedLine[j] = tw.Empty
				}
				prepared[i] = paddedLine
			} else if len(prepared[i]) > numColsBatch {
				t.logger.Debug("Truncating %s line %d from %d to %d columns", sectionName, i, len(prepared[i]), numColsBatch)
				prepared[i] = prepared[i][:numColsBatch]
			}
		}
	}

	return prepared
}

// convertToString converts a value to its string representation.
func (t *Table) convertToString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case tw.Formatter:
		return v.Format()
	case io.Reader:
		const maxReadSize = 512
		var buf strings.Builder
		_, err := io.CopyN(&buf, v, maxReadSize)
		if err != nil && err != io.EOF {
			return fmt.Sprintf("[reader error: %v]", err)
		}
		if buf.Len() == maxReadSize {
			buf.WriteString("...")
		}
		return buf.String()
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return ""
	case sql.NullInt64:
		if v.Valid {
			return fmt.Sprintf("%d", v.Int64)
		}
		return ""
	case sql.NullFloat64:
		if v.Valid {
			return fmt.Sprintf("%f", v.Float64)
		}
		return ""
	case sql.NullBool:
		if v.Valid {
			return fmt.Sprintf("%t", v.Bool)
		}
		return ""
	case sql.NullTime:
		if v.Valid {
			return v.Time.String()
		}
		return ""
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		defer func() {
			if r := recover(); r != nil {
				t.logger.Debug("convertToString: Recovered panic for value %v: %v", value, r)
			}
		}()
		return fmt.Sprintf("%v", value)
	}
}

// callStringer invokes the table's stringer function with optional caching.
func (t *Table) callStringer(input interface{}) ([]string, error) {
	if !t.stringerCacheEnabled {
		// Fallback to original reflection logic
		rv := reflect.ValueOf(t.stringer)
		stringerType := rv.Type()
		inputValue := reflect.ValueOf(input)

		if !(rv.Kind() == reflect.Func && stringerType.NumIn() == 1 && stringerType.NumOut() == 1 &&
			inputValue.Type().AssignableTo(stringerType.In(0)) &&
			stringerType.Out(0) == reflect.TypeOf([]string{})) {
			return nil, errors.Newf("stringer must be func(T) []string where T is assignable from %T, got %T", input, t.stringer)
		}
		result := rv.Call([]reflect.Value{inputValue})
		cells, ok := result[0].Interface().([]string)
		if !ok {
			return nil, errors.Newf("stringer for type %T did not return []string", input)
		}
		return cells, nil
	}

	// Cached path
	inputType := reflect.TypeOf(input)
	t.stringerCacheMu.RLock()
	if rv, ok := t.stringerCache[inputType]; ok {
		t.stringerCacheMu.RUnlock()
		t.logger.Debug("callStringer: Cache hit for type %v", inputType)
		result := rv.Call([]reflect.Value{reflect.ValueOf(input)})
		cells, ok := result[0].Interface().([]string)
		if !ok {
			return nil, errors.Newf("cached stringer for type %T did not return []string", input)
		}
		return cells, nil
	}
	t.stringerCacheMu.RUnlock()

	t.logger.Debug("callStringer: Cache miss for type %v", inputType)
	rv := reflect.ValueOf(t.stringer)
	stringerType := rv.Type()
	if !(rv.Kind() == reflect.Func && stringerType.NumIn() == 1 && stringerType.NumOut() == 1 &&
		inputType.AssignableTo(stringerType.In(0)) &&
		stringerType.Out(0) == reflect.TypeOf([]string{})) {
		return nil, errors.Newf("stringer must be func(T) []string where T is assignable from %T, got %T", input, t.stringer)
	}

	t.stringerCacheMu.Lock()
	t.stringerCache[inputType] = rv
	t.stringerCacheMu.Unlock()

	result := rv.Call([]reflect.Value{reflect.ValueOf(input)})
	cells, ok := result[0].Interface().([]string)
	if !ok {
		return nil, errors.Newf("stringer for type %T did not return []string", input)
	}
	return cells, nil
}

// buildPaddingLineContents constructs a padding line for a given section, respecting column widths and horizontal merges.
// It generates a []string where each element is the padding content for a column, using the specified padChar.
func (t *Table) buildPaddingLineContents(padChar string, widths tw.Mapper[int, int], numCols int, merges map[int]tw.MergeState) []string {
	line := make([]string, numCols)
	padWidth := twfn.DisplayWidth(padChar)
	if padWidth < 1 {
		padWidth = 1
	}
	for j := 0; j < numCols; j++ {
		mergeState := tw.MergeState{}
		if merges != nil {
			if state, ok := merges[j]; ok {
				mergeState = state
			}
		}
		if mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
			line[j] = tw.Empty
			continue
		}
		colWd := widths.Get(j)
		repeatCount := 0
		if colWd > 0 && padWidth > 0 {
			repeatCount = colWd / padWidth
		}
		if colWd > 0 && repeatCount < 1 {
			repeatCount = 1
		}
		content := strings.Repeat(padChar, repeatCount)
		line[j] = content
	}
	if t.logger.Enabled() {
		t.logger.Debug("Built padding line with char '%s' for %d columns", padChar, numCols)
	}
	return line
}
