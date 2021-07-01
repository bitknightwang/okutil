package okutil

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func PrintJson(o interface{}) {
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		Warnf("%v", o)
	} else {
		Info("\n", string(b))
	}
}

func DebugJson(o interface{}) {
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		Debugf("%v", o)
	} else {
		Debug("\n", string(b))
	}
}

func TraceJson(o interface{}) {
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		Tracef("%v", o)
	} else {
		Trace("\n", string(b))
	}
}

func PrintTable(o interface{}) {
	data, header := AnatomyInput(o, 1)
	OutputTable(header, data, nil)
}

func AnatomyInput(o interface{}, nestLevel int) (data [][]string, header []string) {
	ot := reflect.TypeOf(o)
	ov := reflect.ValueOf(o)

	var row []string
	switch ot.Kind() {
	case reflect.Map:
		Debugf("%v is a map", o)
		for _, key := range ov.MapKeys() {
			val := ov.MapIndex(key)
			Debugf("map[%v]:%v = value[%v]:%v\n", key, key.Type(), val, val.Type())

			header = append(header, fmt.Sprintf("%v", key))
			row = append(row, fmt.Sprintf("%v", val))
		}
		data = append(data, row)
	case reflect.Array, reflect.Slice:
		Debugf("%v is a array or slice", o)
		if nestLevel > 0 {
			if ov.Len() > 0 {
				for i := 0; i < ov.Len(); i++ {
					// nest
					Debugf("elem[%v]:%v = %v", i, ov.Index(i).Type(), ov.Index(i).Interface())
					nestedData, nestedHeader := AnatomyInput(ov.Index(i).Interface(), nestLevel-1)
					if len(header) == 0 {
						header = nestedHeader
					}
					for _, row = range nestedData {
						data = append(data, row)
					}
				}
			}
		} else {
			row = append(row, fmt.Sprintf("%v", o))
			data = append(data, row)
		}
	case reflect.Struct:
		Debugf("%v is a struct", o)
		for i := 0; i < ot.NumField(); i++ {
			f := ot.Field(i)
			Debugf("field[%v]:%v = value[%v]:%v\n", f.Name, f.Type, ov.Field(i).Interface(), ov.Field(i).Type())

			header = append(header, fmt.Sprintf("%v", f.Name))
			row = append(row, fmt.Sprintf("%v", ov.Field(i).Interface()))
		}
		data = append(data, row)
	default:
		Debugf("%v:%v = %v", ot.Name(), ot, ov.Interface())

		header = append(header, fmt.Sprintf("%v", ot.Name()))
		row = append(row, fmt.Sprintf("%v", ov.Interface()))
		data = append(data, row)
	}

	return
}

func OutputTable(header []string, data [][]string, footer []string) {
	table := tablewriter.NewWriter(os.Stdout)

	if len(header) > 0 {
		table.SetHeader(header)
	}

	if len(footer) > 0 {
		table.SetFooter(footer)
	}

	table.SetBorder(true)
	table.SetRowLine(true)

	table.AppendBulk(data)

	table.Render()
}

func escapeDelimiter(col string, delimiter string) string {
	if len(col) > 0 && len(delimiter) == 1 {
		escaped := ""
		for _, c := range col {
			if string(c) == delimiter {
				escaped = escaped + "\\" + string(c)
			} else {
				escaped = escaped + string(c)
			}
		}
		return escaped
	}
	return col
}

func PrintCSV(o interface{}, fieldSep string, around string) {
	var sep string
	if len(fieldSep) == 0 {
		sep = ","
	} else {
		sep = fieldSep
	}

	data, header := AnatomyInput(o, 1)

	// FIXME: escape around char in column
	body := "\n"
	if len(header) > 0 {
		if len(around) > 0 {
			var formattedHeader []string
			for _, col := range header {
				formattedHeader = append(formattedHeader, fmt.Sprintf("%v%v%v", around, escapeDelimiter(col, around), around))
			}
			body = body + fmt.Sprintf("%v", strings.Join(formattedHeader, sep))
		} else {
			body = body + fmt.Sprintf("%v", strings.Join(header, sep))
		}
	}

	for _, row := range data {
		if len(around) > 0 {
			var formattedRow []string
			for _, col := range row {
				formattedRow = append(formattedRow, fmt.Sprintf("%v%v%v", around, escapeDelimiter(col, around), around))
			}
			body = body + fmt.Sprintf("\n%v", strings.Join(formattedRow, sep))
		} else {
			body = body + fmt.Sprintf("\n%v", strings.Join(row, sep))
		}
	}

	Info(body)
}
