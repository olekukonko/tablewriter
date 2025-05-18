## More Examples 

> NOTE
Extracted form Issues 

### 1. Merging 

```go
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"os"
)

func main() {
	data := [][]string{
		{"Package", "Version", "Status"},
		{"table\nwriter", "v0.0.1", "legacy"},
		{"table\nwriter", "v0.0.2", "legacy"},
		{"table\nwriter", "v0.0.2", "legacy"},
		{"table\nwriter", "v0.0.2", "legacy"},
		{"table\nwriter", "v0.0.5", "legacy"},
		{"table\nwriter", "v1.0.6", "latest"},
	}

	r := tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{BetweenRows: tw.On},
			Lines:      tw.Lines{ShowFooterLine: tw.On},
		},
	}))
	table := tablewriter.NewTable(os.Stdout, r,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode:  tw.MergeHierarchical,
					Alignment:  tw.AlignCenter,
					AutoWrap:   tw.WrapNone,
					AutoFormat: tw.Off,
				},
			},
		}),
	)

	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()

	table = tablewriter.NewTable(os.Stdout, r,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode: tw.MergeVertical,
					Alignment: tw.AlignCenter,
					AutoWrap:  tw.WrapNone,
				},
			},
		}),
	)

	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()

}

```


```
┌─────────┬─────────┬────────┐
│ PACKAGE │ VERSION │ STATUS │
├─────────┼─────────┼────────┤
│  table  │ v0.0.1  │ legacy │
│ writer  │         │        │
│         ├─────────┼────────┤
│         │ v0.0.2  │ legacy │
│         │         │        │
│         │         │        │
│         │         │        │
│         │         │        │
│         ├─────────┼────────┤
│         │ v0.0.5  │ legacy │
│         ├─────────┼────────┤
│         │ v1.0.6  │ latest │
└─────────┴─────────┴────────┘
┌─────────┬─────────┬────────┐
│ PACKAGE │ VERSION │ STATUS │
├─────────┼─────────┼────────┤
│  table  │ v0.0.1  │ legacy │
│ writer  │         │        │
│         ├─────────┤        │
│         │ v0.0.2  │        │
│         │         │        │
│         │         │        │
│         │         │        │
│         │         │        │
│         ├─────────┤        │
│         │ v0.0.5  │        │
│         ├─────────┼────────┤
│         │ v1.0.6  │ latest │
└─────────┴─────────┴────────┘
```

### 2. Kubectl style output

```go
package main

import (
	"os"
	"sync"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

var wg sync.WaitGroup

func main() {
	data := [][]string{
		{"node1.example.com", "Ready", "compute", "1.11"},
		{"node2.example.com", "Ready", "compute", "1.11"},
		{"node3.example.com", "Ready", "compute", "1.11"},
		{"node4.example.com", "NotReady", "compute", "1.11"},
	}

	table := tablewriter.NewTable(os.Stdout,

		// tell render not to render any lines and separators
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.SeparatorsNone,
				Lines:      tw.LinesNone,
			},
		})),

		// Set general configuration
		tablewriter.WithConfig(
			tablewriter.Config{
				Header: tw.CellConfig{
					Formatting: tw.CellFormatting{
						Alignment: tw.AlignLeft, // force alignment for header
					},
				},
				Row: tw.CellConfig{
					Formatting: tw.CellFormatting{
						Alignment: tw.AlignLeft, // force alightment for body
					},

					// remove all padding in a in all cells
					Padding: tw.CellPadding{Global: tw.PaddingNone},
				},
			},
		),
	)
	table.Header("Name", "Status", "Role", "Version")
	table.Bulk(data)
	table.Render()
}

```

```go
 NAME               STATUS    ROLE     VERSION 
 node1.example.com  Ready     compute  1.11    
 node2.example.com  Ready     compute  1.11    
 node3.example.com  Ready     compute  1.11    
 node4.example.com  NotReady  compute  1.11
```