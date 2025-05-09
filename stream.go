package tablewriter

import (
	"fmt"
	"github.com/olekukonko/errors"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"math"
	"strings"
)

// Start initializes the table stream.
// In this streaming model, renderer.Start() is primarily called in NewStreamTable.
// This method serves as a safeguard or point for adding pre-rendering logic.
// Start initializes the table stream.
// It is the entry point for streaming mode.
// Requires t.config.Stream.Enable to be true.
// Returns an error if streaming is disabled or the renderer does not support streaming,
// or if called multiple times on the same stream.
func (t *Table) Start() error {
	t.ensureInitialized() // Ensures basic setup like loggers

	// --- MODIFIED LOGIC START ---
	if !t.config.Stream.Enable {
		// Start() should only be called when streaming is explicitly enabled.
		// Otherwise, the user should call Render() for batch mode.
		t.logger.Warn("Start() called but streaming is disabled. Call Render() instead for batch mode.")
		return errors.New("start() called but streaming is disabled")
	}

	if !t.renderer.Config().Streaming {
		// Check if the configured renderer actually supports streaming.
		t.logger.Error("Configured renderer does not support streaming.")
		return fmt.Errorf("renderer does not support streaming")
	}

	//t.renderer.Start(t.writer)
	//t.renderer.Logger(t.logger)

	if t.hasPrinted {
		// Prevent calling Start() multiple times on the same stream instance.
		t.logger.Warn("Start() called multiple times for the same table stream. Ignoring subsequent calls.")
		return nil
	}

	t.logger.Debug("Starting table stream.")

	// Initialize/reset streaming state flags and buffers
	t.headerRendered = false
	t.firstRowRendered = false
	t.lastRenderedLineContent = nil
	t.lastRenderedPosition = "" // Reset last rendered position
	t.streamFooterLines = nil   // Reset footer buffer
	t.streamNumCols = 0         // Reset derived column count

	// Calculate initial fixed widths if provided in StreamConfig.Widths
	// These widths will be used for all subsequent rendering in streaming mode.
	if t.config.Stream.Widths.PerColumn != nil && t.config.Stream.Widths.PerColumn.Len() > 0 {
		// Use per-column stream widths if set
		t.logger.Debug("Using per-column stream widths from StreamConfig: %v", t.config.Stream.Widths.PerColumn)
		t.streamWidths = t.config.Stream.Widths.PerColumn.Clone()
		// Determine numCols from the highest index in PerColumn map
		maxColIdx := -1
		t.streamWidths.Each(func(col int, width int) {
			if col > maxColIdx {
				maxColIdx = col
			}
			// Ensure configured widths are reasonable (>0 becomes >=1, <0 becomes 0)
			if width > 0 && width < 1 {
				t.streamWidths.Set(col, 1)
			} else if width < 0 {
				t.streamWidths.Set(col, 0) // Negative width means hide column
			}
		})
		if maxColIdx >= 0 {
			t.streamNumCols = maxColIdx + 1
			t.logger.Debug("Derived streamNumCols from PerColumn widths: %d", t.streamNumCols)
		} else {
			// PerColumn map exists but is empty? Or all negative widths? Assume 0 columns for now.
			t.streamNumCols = 0
			t.logger.Debug("PerColumn widths map is effectively empty or contains only negative values, streamNumCols = 0.")
		}

	} else if t.config.Stream.Widths.Global > 0 {
		// Global width is set, but we don't know the number of columns yet.
		// Defer applying global width until the first data (Header or first Row) arrives.
		// Store a placeholder or flag indicating global width should be used.
		// The simple way for now: Keep streamWidths empty, signal the global width preference.
		// The width calculation function called later will need to check StreamConfig.Widths.Global
		// if streamWidths is empty.
		t.logger.Debug("Global stream width %d set in StreamConfig. Will derive numCols from first data.", t.config.Stream.Widths.Global)
		t.streamWidths = tw.NewMapper[int, int]() // Initialize as empty, will be populated later
		// Note: No need to store Global width value here, it's available in t.config.Stream.Widths.Global

	} else {
		// No explicit stream widths in config. They will be calculated from the first data (Header or first Row).
		t.logger.Debug("No explicit stream widths configured in StreamConfig. Will derive from first data.")
		t.streamWidths = tw.NewMapper[int, int]() // Initialize as empty, will be populated later
		t.streamNumCols = 0                       // NumCols will be determined by first data
	}

	// Log warnings if incompatible features are enabled in streaming config
	// Vertical/Hierarchical merges require processing all rows together.
	if t.config.Header.Formatting.MergeMode&(tw.MergeVertical|tw.MergeHierarchical) != 0 {
		t.logger.Warn("Vertical or Hierarchical merge modes enabled on Header config (%d) but are unsupported in streaming mode. Only Horizontal merge will be considered.", t.config.Header.Formatting.MergeMode)
	}
	if t.config.Row.Formatting.MergeMode&(tw.MergeVertical|tw.MergeHierarchical) != 0 {
		t.logger.Warn("Vertical or Hierarchical merge modes enabled on Row config (%d) but are unsupported in streaming mode. Only Horizontal merge will be considered.", t.config.Row.Formatting.MergeMode)
	}
	if t.config.Footer.Formatting.MergeMode&(tw.MergeVertical|tw.MergeHierarchical) != 0 {
		t.logger.Warn("Vertical or Hierarchical merge modes enabled on Footer config (%d) but are unsupported in streaming mode. Only Horizontal merge will be considered.", t.config.Footer.Formatting.MergeMode)
	}
	// AutoHide requires processing all row data to find empty columns.
	if t.config.AutoHide {
		t.logger.Warn("AutoHide is enabled in config but is ignored in streaming mode.")
	}

	// Call the renderer's start method for the stream.
	err := t.renderer.Start(t.writer)
	if err == nil {
		t.hasPrinted = true // Mark as started successfully only if renderer.Start works
		t.logger.Debug("Renderer.Start() succeeded. Table stream initiated.")
	} else {
		// Reset state if renderer.Start fails
		t.hasPrinted = false
		t.headerRendered = false
		t.firstRowRendered = false
		t.lastRenderedLineContent = nil
		t.lastRenderedPosition = ""
		t.streamFooterLines = nil
		t.streamWidths = tw.NewMapper[int, int]() // Clear any widths that might have been set
		t.streamNumCols = 0
		t.logger.Error("Renderer.Start() failed: %v. Streaming initialization failed.", err)
	}
	return err
}

// Close finalizes the table stream.
// It requires the stream to be started (by calling NewStreamTable).
// It calls the renderer's Close method to render final elements (like the bottom border) and close the stream.
func (t *Table) Close() error {
	t.logger.Debug("Close() called. Finalizing stream.")

	// Ensure stream was actually started and enabled
	if !t.config.Stream.Enable || !t.hasPrinted {
		t.logger.Warn("Close() called but streaming not enabled or not started. Ignoring Close() actions.")
		// If renderer has a Close method that should always be called, consider that.
		// For Blueprint, Close is a no-op, so returning early is fine.
		// If we always call renderer.Close(), ensure it's safe if renderer.Start() wasn't called.
		// Let's only call renderer.Close if stream was started.
		if t.hasPrinted && t.renderer != nil { // Check if renderer is not nil for safety
			t.renderer.Close(t.writer) // Still call renderer's close for cleanup
		}
		t.hasPrinted = false // Reset flag
		return nil
	}

	// Render stored footer if any
	if len(t.streamFooterLines) > 0 {
		t.logger.Debug("Close(): Rendering stored footer.")
		if err := t.streamRenderFooter(t.streamFooterLines); err != nil {
			t.logger.Error("Close(): Failed to render stream footer: %v", err)
			// Continue to try and close renderer and render bottom border
		}
	}

	// Render the final table bottom border
	t.logger.Debug("Close(): Rendering stream bottom border.")
	if err := t.renderStreamBottomBorder(); err != nil {
		t.logger.Error("Close(): Failed to render stream bottom border: %v", err)
		// Continue to try and close renderer
	}

	// Call the underlying renderer's Close method
	err := t.renderer.Close(t.writer)
	if err != nil {
		t.logger.Error("Renderer.Close() failed: %v", err)
	}

	// Reset streaming state
	t.hasPrinted = false
	t.headerRendered = false
	t.firstRowRendered = false
	t.lastRenderedLineContent = nil
	t.lastRenderedMergeState = nil
	t.lastRenderedPosition = ""
	t.streamFooterLines = nil
	// t.streamWidths should persist if we want to make multiple Start/Close calls on same config?
	// For now, let's assume Start re-evaluates. If widths are from StreamConfig, they'd be reused.
	// If derived, they'd be re-derived. Let's clear for true reset.
	t.streamWidths = tw.NewMapper[int, int]()
	t.streamNumCols = 0
	// t.streamRowCounter = 0 // Removed this field

	t.logger.Debug("Stream ended. hasPrinted = false.")
	return err // Return error from renderer.Close or other significant errors
}

// calculateStreamWidths determines the fixed column widths for streaming mode.
// It prioritizes widths from StreamConfig.Widths.PerColumn, then StreamConfig.Widths.Global,
// then derives from the provided sample data lines.
// It populates t.streamWidths and t.streamNumCols if they are currently empty.
// The sampleDataLines should be the *raw* input lines (e.g., []string for Header/Footer, or the first row's []string cells for Row).
// The paddingConfig should be the CellPadding config relevant to the sample data (Header/Row/Footer).
// Returns the determined number of columns.
// This function should only be called when t.streamWidths is currently empty.
func (t *Table) calculateStreamWidths(sampleDataLines []string, sectionConfigForSampleData tw.CellConfig) int {
	if t.streamWidths != nil && t.streamWidths.Len() > 0 {
		t.logger.Debug("calculateStreamWidths: Called when streaming widths are already set (%d columns). Reusing existing.", t.streamNumCols)
		return t.streamNumCols
	}

	t.logger.Debug("calculateStreamWidths: Calculating streaming widths. Sample data cells: %d. Using section config: %+v", len(sampleDataLines), sectionConfigForSampleData.Formatting)

	determinedNumCols := 0
	if t.config.Stream.Widths.PerColumn != nil && t.config.Stream.Widths.PerColumn.Len() > 0 {
		maxColIdx := -1
		t.config.Stream.Widths.PerColumn.Each(func(col int, width int) {
			if col > maxColIdx {
				maxColIdx = col
			}
		})
		determinedNumCols = maxColIdx + 1
		t.logger.Debug("calculateStreamWidths: Determined numCols (%d) from StreamConfig.Widths.PerColumn", determinedNumCols)
	} else if len(sampleDataLines) > 0 {
		determinedNumCols = len(sampleDataLines)
		t.logger.Debug("calculateStreamWidths: Determined numCols (%d) from sample data length", determinedNumCols)
	} else {
		t.logger.Debug("calculateStreamWidths: Cannot determine numCols (no PerColumn config, no sample data)")
		t.streamNumCols = 0
		t.streamWidths = tw.NewMapper[int, int]()
		return 0
	}

	t.streamNumCols = determinedNumCols
	t.streamWidths = tw.NewMapper[int, int]()

	// Use padding and autowrap from the provided sectionConfigForSampleData
	paddingForWidthCalc := sectionConfigForSampleData.Padding
	autoWrapForWidthCalc := sectionConfigForSampleData.Formatting.AutoWrap

	if t.config.Stream.Widths.PerColumn != nil && t.config.Stream.Widths.PerColumn.Len() > 0 {
		t.logger.Debug("calculateStreamWidths: Using widths from StreamConfig.Widths.PerColumn")
		for i := 0; i < t.streamNumCols; i++ {
			width, ok := t.config.Stream.Widths.PerColumn.OK(i)
			if !ok {
				width = 0
			}
			if width > 0 && width < 1 {
				width = 1
			} else if width < 0 {
				width = 0
			}
			t.streamWidths.Set(i, width)
		}
	} else {
		// No PerColumn config, derive from sampleDataLines intelligently
		t.logger.Debug("calculateStreamWidths: Intelligently deriving widths from sample data content and padding.")
		tempRequiredWidths := tw.NewMapper[int, int]() // Widths from updateWidths (content + padding)
		if len(sampleDataLines) > 0 {
			// updateWidths calculates: DisplayWidth(content) + padLeft + padRight
			t.updateWidths(sampleDataLines, tempRequiredWidths, paddingForWidthCalc)
		}

		ellipsisWidthBuffer := 0
		if autoWrapForWidthCalc == tw.WrapTruncate {
			ellipsisWidthBuffer = twfn.DisplayWidth(tw.CharEllipsis)
		}
		varianceBuffer := 2 // Your suggested variance
		minTotalColWidth := tw.DefaultMinlColumnWidth
		// Example: if t.config.Stream.MinAutoColumnWidth > 0 { minTotalColWidth = t.config.Stream.MinAutoColumnWidth }

		for i := 0; i < t.streamNumCols; i++ {
			// baseCellWidth (content_width + padding_width) comes from tempRequiredWidths.Get(i)
			// We need to deconstruct it to apply logic to content_width first.

			sampleContent := ""
			if i < len(sampleDataLines) {
				sampleContent = strings.TrimSpace(sampleDataLines[i])
			}
			sampleContentDisplayWidth := twfn.DisplayWidth(sampleContent)

			colPad := paddingForWidthCalc.Global
			if i < len(paddingForWidthCalc.PerColumn) && paddingForWidthCalc.PerColumn[i] != (tw.Padding{}) {
				colPad = paddingForWidthCalc.PerColumn[i]
			}
			currentPadLWidth := twfn.DisplayWidth(colPad.Left)
			currentPadRWidth := twfn.DisplayWidth(colPad.Right)
			currentTotalPaddingWidth := currentPadLWidth + currentPadRWidth

			// Start with the target content width logic
			targetContentWidth := sampleContentDisplayWidth
			if autoWrapForWidthCalc == tw.WrapTruncate {
				// If content is short, ensure it's at least wide enough for an ellipsis
				if targetContentWidth < ellipsisWidthBuffer {
					targetContentWidth = ellipsisWidthBuffer
				}
			}
			targetContentWidth += varianceBuffer // Add variance

			// Now calculate the total cell width based on this buffered content target + padding
			calculatedWidth := targetContentWidth + currentTotalPaddingWidth

			// Apply an absolute minimum total column width
			if calculatedWidth > 0 && calculatedWidth < minTotalColWidth {
				t.logger.Debug("calculateStreamWidths: Col %d, InitialCalcW=%d (ContentTarget=%d + Pad=%d) is less than MinTotalW=%d. Adjusting to MinTotalW.",
					i, calculatedWidth, targetContentWidth, currentTotalPaddingWidth, minTotalColWidth)
				calculatedWidth = minTotalColWidth
			} else if calculatedWidth <= 0 && sampleContentDisplayWidth > 0 { // If content exists but calc width is 0 (e.g. large negative variance)
				// Ensure at least min width or content + padding + buffers
				fallbackWidth := sampleContentDisplayWidth + currentTotalPaddingWidth
				if autoWrapForWidthCalc == tw.WrapTruncate {
					fallbackWidth += ellipsisWidthBuffer
				}
				fallbackWidth += varianceBuffer
				calculatedWidth = twfn.Max(minTotalColWidth, fallbackWidth)
				if calculatedWidth <= 0 && (currentTotalPaddingWidth+1) > 0 { // last resort if all else is zero
					calculatedWidth = currentTotalPaddingWidth + 1
				} else if calculatedWidth <= 0 {
					calculatedWidth = 1 // absolute last resort
				}

				t.logger.Debug("calculateStreamWidths: Col %d, CalculatedW was <=0 despite content. Adjusted to %d.", i, calculatedWidth)
			} else if calculatedWidth <= 0 && sampleContentDisplayWidth == 0 {
				// Column is truly empty in sample and buffers didn't make it positive, or minTotalColWidth is 0.
				// Keep width 0 (it will be hidden by renderer if all content is empty for this col)
				// Or, if we want empty columns to have a minimum presence (even if just padding):
				// calculatedWidth = currentTotalPaddingWidth // This would make it just wide enough for padding
				// For now, let truly empty sample + no min width result in 0.
				calculatedWidth = 0 // Explicitly set to 0 if it ended up non-positive and no content
			}

			t.streamWidths.Set(i, calculatedWidth)
			t.logger.Debug("calculateStreamWidths: Col %d, SampleContentW=%d, PadW=%d, EllipsisBufIfTruncate=%d, VarianceBuf=%d -> FinalTotalColW=%d",
				i, sampleContentDisplayWidth, currentTotalPaddingWidth, ellipsisWidthBuffer, varianceBuffer, calculatedWidth)
		}
	}

	// Apply Global Constraint (if t.config.Stream.Widths.Global > 0)
	if t.config.Stream.Widths.Global > 0 && t.streamNumCols > 0 {
		t.logger.Debug("calculateStreamWidths: Applying global stream width constraint %d", t.config.Stream.Widths.Global)
		currentTotalColumnWidthsSum := 0
		t.streamWidths.Each(func(_ int, w int) {
			currentTotalColumnWidthsSum += w
		})

		separatorWidth := 0
		if t.renderer != nil {
			rendererConfig := t.renderer.Config()
			if rendererConfig.Settings.Separators.BetweenColumns.Enabled() {
				separatorWidth = twfn.DisplayWidth(rendererConfig.Symbols.Column())
			}
		} else {
			separatorWidth = 1 // Default if renderer not available yet
		}

		totalWidthIncludingSeparators := currentTotalColumnWidthsSum
		if t.streamNumCols > 1 {
			totalWidthIncludingSeparators += (t.streamNumCols - 1) * separatorWidth
		}

		if t.config.Stream.Widths.Global < totalWidthIncludingSeparators && totalWidthIncludingSeparators > 0 { // Added check for total > 0
			t.logger.Debug("calculateStreamWidths: Total calculated width (%d incl separators) exceeds global stream width (%d). Shrinking.", totalWidthIncludingSeparators, t.config.Stream.Widths.Global)

			// Target sum for column widths only (global limit - total separator width)
			targetSumForColumnWidths := t.config.Stream.Widths.Global
			if t.streamNumCols > 1 {
				targetSumForColumnWidths -= (t.streamNumCols - 1) * separatorWidth
			}
			if targetSumForColumnWidths < t.streamNumCols && t.streamNumCols > 0 { // Ensure at least 1 per column if possible
				targetSumForColumnWidths = t.streamNumCols
			} else if targetSumForColumnWidths < 0 {
				targetSumForColumnWidths = 0
			}

			scaleFactor := float64(targetSumForColumnWidths) / float64(currentTotalColumnWidthsSum)
			if currentTotalColumnWidthsSum <= 0 {
				scaleFactor = 0
			} // Avoid division by zero or negative scale

			adjustedSum := 0
			for i := 0; i < t.streamNumCols; i++ {
				originalColWidth := t.streamWidths.Get(i)
				if originalColWidth == 0 {
					continue
				} // Don't scale hidden columns

				scaledWidth := 0
				if scaleFactor > 0 {
					scaledWidth = int(math.Round(float64(originalColWidth) * scaleFactor))
				}

				if scaledWidth < 1 && originalColWidth > 0 { // Ensure at least 1 if original had width and scaling made it too small
					scaledWidth = 1
				} else if scaledWidth < 0 { // Should not happen with math.Round on positive*positive
					scaledWidth = 0
				}
				t.streamWidths.Set(i, scaledWidth)
				adjustedSum += scaledWidth
			}

			// Distribute rounding errors to meet targetSumForColumnWidths
			remainingSpace := targetSumForColumnWidths - adjustedSum
			t.logger.Debug("calculateStreamWidths: Scaling complete. TargetSum=%d, AchievedSum=%d, RemSpace=%d", targetSumForColumnWidths, adjustedSum, remainingSpace)
			// Distribute remainingSpace (positive or negative) among non-zero width columns
			if remainingSpace != 0 && t.streamNumCols > 0 {
				colsToAdjust := []int{}
				t.streamWidths.Each(func(col int, w int) {
					if w > 0 { // Only consider columns that currently have width
						colsToAdjust = append(colsToAdjust, col)
					}
				})
				if len(colsToAdjust) > 0 {
					for i := 0; i < int(math.Abs(float64(remainingSpace))); i++ {
						colIdx := colsToAdjust[i%len(colsToAdjust)]
						currentColWidth := t.streamWidths.Get(colIdx)
						if remainingSpace > 0 {
							t.streamWidths.Set(colIdx, currentColWidth+1)
						} else if remainingSpace < 0 && currentColWidth > 1 { // Don't reduce below 1
							t.streamWidths.Set(colIdx, currentColWidth-1)
						}
					}
				}
			}
			t.logger.Debug("calculateStreamWidths: Widths after scaling and distribution: %v", t.streamWidths)
		} else {
			t.logger.Debug("calculateStreamWidths: Total calculated width (%d) fits global stream width (%d). No scaling needed.", totalWidthIncludingSeparators, t.config.Stream.Widths.Global)
		}
	}

	// Final sanitization
	t.streamWidths.Each(func(col int, width int) {
		if width < 0 {
			t.streamWidths.Set(col, 0)
		}
	})

	t.logger.Debug("calculateStreamWidths: Final derived stream widths after all adjustments (%d columns): %v", t.streamNumCols, t.streamWidths)
	return t.streamNumCols
}

// streamRenderHeader processes and renders the header section in streaming mode.
// It calculates/uses fixed stream widths, processes content, renders borders/lines,
// and updates streaming state.
// It assumes Start() has already been called and t.hasPrinted is true.
func (t *Table) streamRenderHeader(headers []string) error {
	t.logger.Debug("streamRenderHeader called with headers: %v", headers)

	if t.headerRendered {
		t.logger.Warn("streamRenderHeader called but header already rendered. Ignoring.")
		return nil
	}

	if t.streamWidths == nil || t.streamWidths.Len() == 0 {
		t.logger.Debug("streamRenderHeader: Stream widths not set, calculating from header data and config.")
		t.calculateStreamWidths(headers, t.config.Header)
		if t.streamNumCols == 0 {
			t.logger.Warn("streamRenderHeader: Failed to determine column count. Cannot render header.")
			return errors.New("failed to determine column count for streaming header")
		}
		for i := 0; i < t.streamNumCols; i++ {
			if _, ok := t.streamWidths.OK(i); !ok {
				t.streamWidths.Set(i, 0)
			}
		}
		t.logger.Debug("streamRenderHeader: Determined stream widths: %v", t.streamWidths)
	} else {
		t.logger.Debug("streamRenderHeader: Stream widths already set (%d cols): %v", t.streamNumCols, t.streamWidths)
		if len(headers) != t.streamNumCols {
			t.logger.Warn("streamRenderHeader: Input header col count (%d) != stream col count (%d). Padding/truncating.", len(headers), t.streamNumCols)
			if len(headers) < t.streamNumCols {
				paddedHeaders := make([]string, t.streamNumCols)
				copy(paddedHeaders, headers)
				for i := len(headers); i < t.streamNumCols; i++ {
					paddedHeaders[i] = tw.Empty
				}
				headers = paddedHeaders
			} else {
				headers = headers[:t.streamNumCols]
			}
		}
	}

	processedHeaderLines := t.prepareContent(headers, t.config.Header)
	t.logger.Debug("streamRenderHeader: Processed header lines: %d", len(processedHeaderLines))

	if t.streamNumCols > 0 { // Mark header as "processed" even if no content lines, if widths are set
		t.headerRendered = true
	}
	if len(processedHeaderLines) == 0 && t.streamNumCols == 0 { // No widths, no content, truly nothing to do for header
		t.logger.Debug("streamRenderHeader: No header content and no columns determined.")
		return nil
	}

	_, headerMerges, _ := t.prepareWithMerges([][]string{headers}, t.config.Header, tw.Header)

	f := t.renderer
	cfg := t.renderer.Config() // Renderer config

	// --- ADDED: Render Top Border if it's the very first element ---
	// Check if anything (content or separator) has been rendered before.
	// t.lastRenderedPosition is initially empty string.
	if t.lastRenderedPosition == "" && cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		t.logger.Debug("streamRenderHeader: Rendering table top border.")
		var nextCellsCtx map[int]tw.CellContext
		if len(processedHeaderLines) > 0 {
			// Build context for the first line of the header (which is "Next" for the top border)
			firstHeaderLineResp := t.buildStreamCellContexts(
				tw.Header, 0, 0, processedHeaderLines, headerMerges, t.config.Header,
			)
			nextCellsCtx = firstHeaderLineResp.cells
		}

		f.Line(t.writer, tw.Formatting{
			Row: tw.RowContext{
				Widths:       t.streamWidths,
				ColMaxWidths: tw.CellWidth{PerColumn: t.streamWidths},
				Next:         nextCellsCtx,
				Position:     tw.Header, // Border is conceptually part of the header section
				Location:     tw.LocationFirst,
			},
			Level:            tw.LevelHeader,
			IsSubRow:         false,
			Debug:            t.config.Debug,
			NormalizedWidths: t.streamWidths,
		})
		// Do NOT update t.lastRenderedPosition here, as this is a border, not content.
		// The first content line will set it.
		t.logger.Debug("streamRenderHeader: Top border rendered.")
	}
	// --- END ADDED Top Border ---

	if len(processedHeaderLines) == 0 { // If only widths were set but no actual header content lines
		t.logger.Debug("streamRenderHeader: No processed header content lines to render, but headerRendered is true.")
		// Set last rendered position to indicate header section was "visited" for separator logic
		if t.headerRendered { // Ensure headerRendered was set (meaning widths were determined)
			t.lastRenderedPosition = tw.Header
			t.lastRenderedLineContent = nil // No specific content line
			t.lastRenderedMergeState = nil
		}
		return nil
	}

	totalHeaderLines := len(processedHeaderLines)
	for i := 0; i < totalHeaderLines; i++ {
		resp := t.buildStreamCellContexts(
			tw.Header, 0, i, processedHeaderLines, headerMerges, t.config.Header,
		)

		f.Header(t.writer, [][]string{resp.cellsContent}, tw.Formatting{
			Row: tw.RowContext{
				Widths:       t.streamWidths,
				ColMaxWidths: tw.CellWidth{PerColumn: t.streamWidths},
				Current:      resp.cells,
				Previous:     resp.prevCells,
				Next:         resp.nextCells,
				Position:     tw.Header,
				Location:     resp.location, // From buildStreamCellContexts
			},
			Level:            tw.LevelHeader,
			IsSubRow:         (i > 0),
			Debug:            t.config.Debug,
			NormalizedWidths: t.streamWidths,
		})

		t.lastRenderedLineContent = resp.cellsContent
		t.lastRenderedMergeState = make(map[int]tw.MergeState)
		for colIdx, cellCtx := range resp.cells {
			t.lastRenderedMergeState[colIdx] = cellCtx.Merge
		}
		t.lastRenderedPosition = tw.Header
	}

	t.logger.Debug("streamRenderHeader: Header content rendering completed.")
	return nil
}

// streamAppendRow processes and renders a single row in streaming mode.
// It calculates/uses fixed stream widths, processes content, renders separators and lines,
// and updates streaming state.
// It assumes Start() has already been called and t.hasPrinted is true.
func (t *Table) streamAppendRow(row interface{}) error {
	t.logger.Debug("streamAppendRow called with row: %v (type: %T)", row, row)

	// Convert row interface to raw string cells using the helper
	rawCellsSlice, err := t.rawCellsToStrings(row, t.config.Row)
	if err != nil {
		t.logger.Error("streamAppendRow: Failed to convert row to strings: %v", err)
		return fmt.Errorf("failed to convert row to strings: %w", err)
	}

	// Calculate fixed stream widths if not already set by Start() or Header()
	if t.streamWidths == nil || t.streamWidths.Len() == 0 {
		t.logger.Debug("streamAppendRow: Stream widths not set, calculating from first row data and config.")
		t.calculateStreamWidths(rawCellsSlice, t.config.Row) // Pass raw cells and row padding
		if t.streamNumCols == 0 {
			t.logger.Warn("streamAppendRow: Failed to determine column count from first row data. Cannot render row.")
			return errors.New("failed to determine column count for streaming row")
		}
		for i := 0; i < t.streamNumCols; i++ { // Ensure all columns up to streamNumCols have a width
			if _, ok := t.streamWidths.OK(i); !ok {
				t.streamWidths.Set(i, 0)
			}
		}
		t.logger.Debug("streamAppendRow: Determined stream widths: %v", t.streamWidths)
	} else {
		t.logger.Debug("streamAppendRow: Stream widths already set (%d columns): %v", t.streamNumCols, t.streamWidths)
		if t.streamNumCols > 0 && len(rawCellsSlice) != t.streamNumCols { // Check streamNumCols before comparing length
			t.logger.Warn("streamAppendRow: Input row column count (%d) != stream column count (%d). Padding/Truncating.", len(rawCellsSlice), t.streamNumCols)
			if len(rawCellsSlice) < t.streamNumCols {
				paddedCells := make([]string, t.streamNumCols)
				copy(paddedCells, rawCellsSlice)
				for i := len(rawCellsSlice); i < t.streamNumCols; i++ {
					paddedCells[i] = tw.Empty
				}
				rawCellsSlice = paddedCells
			} else {
				rawCellsSlice = rawCellsSlice[:t.streamNumCols]
			}
		}
	}

	if t.streamNumCols == 0 { // If still 0 after all attempts, cannot render
		t.logger.Warn("streamAppendRow: streamNumCols is 0. Cannot render row.")
		// If this was supposed to be the first row, mark firstRowRendered false or handle state
		return errors.New("cannot render row, column count is zero and could not be determined")
	}

	// Detect horizontal merges for this row
	_, rowMerges, _ := t.prepareWithMerges([][]string{rawCellsSlice}, t.config.Row, tw.Row)

	// Process raw cells into multi-line strings
	processedRowLines := t.prepareContent(rawCellsSlice, t.config.Row)
	t.logger.Debug("streamAppendRow: Processed row lines: %d lines", len(processedRowLines))

	// At the beginning of streamAppendRow, AFTER width calculation, but BEFORE separator/content rendering:
	if !t.headerRendered && !t.firstRowRendered && t.lastRenderedPosition == "" { // If this is truly the first data element overall
		cfg := t.renderer.Config()
		if cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
			t.logger.Debug("streamAppendRow: Rendering table top border (first element is a row).")
			// f.Line call for top border (similar to streamRenderHeader's top border logic)
			// For the top border line, the 'Next' context is the first line of this current row
			var nextCellsCtx map[int]tw.CellContext
			if len(processedRowLines) > 0 { // processedRowLines is available here
				firstRowLineResp := t.buildStreamCellContexts( // Use current row's first line for Next
					tw.Row, 0, 0, processedRowLines, rowMerges, t.config.Row,
				)
				nextCellsCtx = firstRowLineResp.cells
			}
			t.renderer.Line(t.writer, tw.Formatting{
				Row: tw.RowContext{
					Widths:       t.streamWidths,
					ColMaxWidths: tw.CellWidth{PerColumn: t.streamWidths},
					Next:         nextCellsCtx,
					Position:     tw.Row, // Border positioned relative to the row it precedes
					Location:     tw.LocationFirst,
				},
				Level:            tw.LevelHeader, // Top border is always LevelHeader visually
				IsSubRow:         false,
				Debug:            t.config.Debug,
				NormalizedWidths: t.streamWidths,
			})
			// Do NOT update t.lastRenderedPosition here for a border line.
			t.logger.Debug("streamAppendRow: Top border rendered.")
		}
	}
	// ... then proceed to separator logic and row content rendering ...

	if len(processedRowLines) == 0 {
		t.logger.Debug("streamAppendRow: No processed row lines to render for this row.")
		// If this was the first attempted row and it resulted in no content,
		// still mark firstRowRendered as true so subsequent rows know they aren't the absolute first.
		if !t.firstRowRendered {
			t.firstRowRendered = true // A "row attempt" occurred
			t.logger.Debug("streamAppendRow: Marked first row rendered (empty content after processing).")
		}
		return nil
	}

	// --- Render Separator Line ---
	f := t.renderer
	cfg := t.renderer.Config()

	shouldDrawHeaderRowSeparator := t.headerRendered && !t.firstRowRendered && cfg.Settings.Lines.ShowHeaderLine.Enabled()
	shouldDrawRowRowSeparator := t.firstRowRendered && cfg.Settings.Separators.BetweenRows.Enabled()

	// Debug logging for separator conditions
	firstCellForLog := ""
	if len(rawCellsSlice) > 0 {
		firstCellForLog = rawCellsSlice[0]
	}
	t.logger.Debug("streamAppendRow: Separator Pre-Check for row starting with '%s': headerRendered=%v, firstRowRendered=%v, ShowHeaderLine=%v, BetweenRows=%v, lastRenderedPos=%q",
		firstCellForLog,
		t.headerRendered,
		t.firstRowRendered,
		cfg.Settings.Lines.ShowHeaderLine.Enabled(),
		cfg.Settings.Separators.BetweenRows.Enabled(),
		t.lastRenderedPosition)
	t.logger.Debug("streamAppendRow: Separator Decision Flags for row starting with '%s': shouldDrawHeaderRowSeparator=%v, shouldDrawRowRowSeparator=%v",
		firstCellForLog,
		shouldDrawHeaderRowSeparator,
		shouldDrawRowRowSeparator)

	if (shouldDrawHeaderRowSeparator || shouldDrawRowRowSeparator) && t.lastRenderedPosition != tw.Position("separator") {
		t.logger.Debug("streamAppendRow: >>> Entered SEPARATOR RENDERING BLOCK for row starting with '%s'. Drawing Line.", firstCellForLog)

		prevCellsCtx := t.lastRenderedMergeStateToCellContexts(t.lastRenderedLineContent, t.lastRenderedMergeState)

		var nextCellsCtx map[int]tw.CellContext = nil
		if len(processedRowLines) > 0 { // Should always be true if we reached here
			nextCellsCtx = make(map[int]tw.CellContext)
			firstRowLineContent := padLine(processedRowLines[0], t.streamNumCols)
			for j := 0; j < t.streamNumCols; j++ {
				merge := tw.MergeState{}
				if rowMerges != nil {
					if state, ok := rowMerges[j]; ok {
						merge = state
					}
				}
				nextCellsCtx[j] = tw.CellContext{Data: firstRowLineContent[j], Width: t.streamWidths.Get(j), Merge: merge}
			}
		}

		separatorLevel := tw.LevelBody
		separatorPosition := tw.Row // Separator is related to rows
		separatorLocation := tw.LocationMiddle

		f.Line(t.writer, tw.Formatting{
			Row: tw.RowContext{
				Widths:       t.streamWidths,
				ColMaxWidths: tw.CellWidth{PerColumn: t.streamWidths},
				Current:      prevCellsCtx, // Context of the line above the separator
				Previous:     nil,          // No line "before" Current in this specific line-drawing context
				Next:         nextCellsCtx, // Context of the line below the separator (first line of current row)
				Position:     separatorPosition,
				Location:     separatorLocation,
			},
			Level:            separatorLevel,
			IsSubRow:         false,
			Debug:            t.config.Debug,
			NormalizedWidths: t.streamWidths,
		})

		t.lastRenderedPosition = tw.Position("separator")
		t.lastRenderedLineContent = nil // Clear content/merge state after a separator
		t.lastRenderedMergeState = nil
		t.logger.Debug("streamAppendRow: Separator line rendered. Updated lastRenderedPosition to 'separator'")
	} else {
		// Log why separator rendering was skipped
		details := ""
		if !(shouldDrawHeaderRowSeparator || shouldDrawRowRowSeparator) {
			details = "neither header/row nor row/row separator was flagged true"
		} else if t.lastRenderedPosition == tw.Position("separator") {
			details = "lastRenderedPosition is already 'separator'"
		} else {
			details = "an unexpected combination of conditions"
		}
		t.logger.Debug("streamAppendRow: Separator not drawn for row '%s' because %s.", firstCellForLog, details)
	}
	// --- End Render Separator Line ---

	// Iterate through processed row lines and render each one.
	totalRowLines := len(processedRowLines)
	for i := 0; i < totalRowLines; i++ {
		resp := t.buildStreamCellContexts(
			tw.Row,
			0, // rowIdx (not used by current buildStreamCellContexts logic for RowIdx in RowContext)
			i,
			processedRowLines,
			rowMerges, // Merges specific to this row
			t.config.Row,
		)

		f.Row(t.writer, resp.cellsContent, tw.Formatting{
			Row: tw.RowContext{
				Widths:       t.streamWidths,
				ColMaxWidths: tw.CellWidth{PerColumn: t.streamWidths},
				Current:      resp.cells,
				Previous:     resp.prevCells,
				Next:         resp.nextCells,
				Position:     tw.Row,
				Location:     resp.location,
				// RowIdx field removed from tw.RowContext
			},
			Level:            tw.LevelBody,
			IsSubRow:         (i > 0),
			Debug:            t.config.Debug,
			NormalizedWidths: t.streamWidths,
			HasFooter:        len(t.streamFooterLines) > 0,
		})

		t.lastRenderedLineContent = resp.cellsContent
		t.lastRenderedMergeState = make(map[int]tw.MergeState)
		for colIdx, cellCtx := range resp.cells {
			t.lastRenderedMergeState[colIdx] = cellCtx.Merge
		}
		t.lastRenderedPosition = tw.Row
	}

	if !t.firstRowRendered { // This is the first row whose content lines were actually rendered
		t.firstRowRendered = true
		t.logger.Debug("streamAppendRow: Marked first row rendered (after processing content).")
	}

	t.logger.Debug("streamAppendRow: Row processing completed for row starting with '%s'.", firstCellForLog)
	return nil
}

// rawCellsToStrings converts a single row interface{} into a slice of raw strings ([]string).
// It handles []string, []any, and uses the custom stringer if provided.
// It applies global and per-column filters *before* processing into multi-lines.
// Returns the slice of strings and an error if conversion or stringer fails.
// rawCellsToStrings converts a single row interface{} into a slice of raw strings ([]string).
// It handles []string, []any, and uses the custom stringer if provided.
// It applies global and per-column filters *before* processing into multi-lines.
// Returns the slice of strings and an error if conversion or stringer fails.
//func (t *Table) rawCellsToStrings(row interface{}) ([]string, error) {
//	t.logger.Debug("rawCellsToStrings: Converting row to raw string cells: %v (type: %T)", row, row)
//	var cells []string
//
//	switch v := row.(type) {
//	case []string:
//		cells = v
//		t.logger.Debug("rawCellsToStrings: Row is already []string")
//	case []any:
//		t.logger.Debug("rawCellsToStrings: Row is []any, converting elements")
//		cells = make([]string, len(v))
//		for i, element := range v {
//			cells[i] = t.elementToString(element)
//		}
//	default:
//		if t.stringer != nil {
//			t.logger.Debug("rawCellsToStrings: Attempting conversion using custom stringer for type %T", row)
//			rv := reflect.ValueOf(t.stringer)
//			stringerType := rv.Type()
//
//			// Basic validation of stringer function signature
//			if rv.Kind() != reflect.Func || stringerType.NumIn() != 1 || stringerType.NumOut() != 1 ||
//				!reflect.TypeOf(row).AssignableTo(stringerType.In(0)) || // Ensure input type is assignable
//				stringerType.Out(0).Kind() != reflect.Slice || stringerType.Out(0).Elem().Kind() != reflect.String {
//				err := errors.Newf("stringer must be a func(T) []string where T is assignable from %T, got %T returning %s", row, t.stringer, stringerType.Out(0).String())
//				t.logger.Debug("rawCellsToStrings: Stringer format/type error: %v", err)
//				return nil, err
//			}
//
//			in := []reflect.Value{reflect.ValueOf(row)}
//			out := rv.Call(in)
//
//			// --- CORRECTED REFLECTION CALL START ---
//			// out[0] is a reflect.Value
//			// out[0].Interface() gets the actual value as an interface{}
//			// We then assert this interface{} to []string
//			outSlice, ok := out[0].Interface().([]string)
//			// --- CORRECTED REFLECTION CALL END ---
//
//			if !ok {
//				err := errors.Newf("stringer must return []string, got %T", out[0].Interface())
//				t.logger.Debug("rawCellsToStrings: Stringer return type mismatch: %v", err)
//				return nil, err
//			}
//			cells = outSlice
//			t.logger.Debug("rawCellsToStrings: Converted row using stringer: %v", cells)
//		} else {
//			err := errors.Newf("cannot convert row type %T to []string; provide a stringer via WithStringer", row)
//			t.logger.Debug("rawCellsToStrings: Conversion error: %v", err)
//			return nil, err
//		}
//	}
//
//	// Apply filters if any
//	if t.config.Row.Filter.Global != nil {
//		t.logger.Debug("rawCellsToStrings: Applying global filter to cells: %v", cells)
//		cells = t.config.Row.Filter.Global(cells)
//		t.logger.Debug("rawCellsToStrings: Cells after global filter: %v", cells)
//	}
//
//	if len(t.config.Row.Filter.PerColumn) > 0 {
//		t.logger.Debug("rawCellsToStrings: Applying per-column filters to cells")
//		numFilters := len(t.config.Row.Filter.PerColumn)
//		// Apply filters up to the number of filters defined or the number of cells, whichever is smaller
//		limit := numFilters
//		if len(cells) < limit {
//			limit = len(cells)
//		}
//
//		for i := 0; i < limit; i++ {
//			if t.config.Row.Filter.PerColumn[i] != nil {
//				originalCell := cells[i]
//				cells[i] = t.config.Row.Filter.PerColumn[i](cells[i])
//				if cells[i] != originalCell {
//					t.logger.Debug("  rawCellsToStrings: Col %d filter applied: '%s' -> '%s'", i, originalCell, cells[i])
//				}
//			}
//		}
//		// If cells slice is longer than filters, process remaining cells without filters
//		// No-op needed as loop only went up to limit
//	}
//
//	t.logger.Debug("rawCellsToStrings: Conversion and filtering completed, raw cells: %v", cells)
//	return cells, nil
//}

func (t *Table) renderStreamBottomBorder() error {
	if t.streamWidths == nil || t.streamWidths.Len() == 0 {
		t.logger.Debug("renderStreamBottomBorder: No stream widths available, skipping bottom border.")
		return nil
	}

	cfg := t.renderer.Config()
	if !cfg.Borders.Bottom.Enabled() || !cfg.Settings.Lines.ShowBottom.Enabled() {
		t.logger.Debug("renderStreamBottomBorder: Bottom border disabled in config, skipping.")
		return nil
	}

	// The bottom border's "Current" context is the last rendered content line
	currentCells := make(map[int]tw.CellContext)
	if t.lastRenderedLineContent != nil {
		// Use a helper to convert last rendered state to cell contexts
		currentCells = t.lastRenderedMergeStateToCellContexts(t.lastRenderedLineContent, t.lastRenderedMergeState)
	} else {
		// No content was ever rendered, but we might still want a bottom border if a top border was drawn.
		// Create empty cell contexts.
		for i := 0; i < t.streamNumCols; i++ {
			currentCells[i] = tw.CellContext{Width: t.streamWidths.Get(i)}
		}
		t.logger.Debug("renderStreamBottomBorder: No previous content line, creating empty context for bottom border.")
	}

	f := t.renderer
	f.Line(t.writer, tw.Formatting{
		Row: tw.RowContext{
			Widths:       t.streamWidths,
			ColMaxWidths: tw.CellWidth{PerColumn: t.streamWidths},
			Current:      currentCells,           // Context of the line *above* the bottom border
			Previous:     nil,                    // No line before this, relative to the border itself (or use lastRendered's previous?)
			Next:         nil,                    // No line after the bottom border
			Position:     t.lastRenderedPosition, // Position of the content above the border (Row or Footer)
			Location:     tw.LocationEnd,         // This is the absolute end
		},
		Level:            tw.LevelFooter, // Bottom border is LevelFooter
		IsSubRow:         false,
		Debug:            t.config.Debug,
		NormalizedWidths: t.streamWidths,
	})
	t.logger.Debug("renderStreamBottomBorder: Bottom border rendered.")
	return nil
}

// streamRenderFooter renders the stored footer lines in streaming mode.
// It's called by Close(). It renders the Row/Footer separator line first.
func (t *Table) streamRenderFooter(processedFooterLines [][]string) error {
	t.logger.Debug("streamRenderFooter: Rendering %d processed footer lines.", len(processedFooterLines))

	if t.streamWidths == nil || t.streamWidths.Len() == 0 || t.streamNumCols == 0 {
		t.logger.Warn("streamRenderFooter: No stream widths or columns defined. Cannot render footer.")
		return errors.New("cannot render stream footer without defined column widths")
	}

	if len(processedFooterLines) == 0 {
		t.logger.Debug("streamRenderFooter: No footer lines to render.")
		return nil
	}

	f := t.renderer
	cfg := t.renderer.Config()

	// --- Render Row/Footer or Header/Footer Separator Line ---
	// This separator is drawn if ShowFooterLine is enabled AND there was content before the footer.
	// The last rendered position (t.lastRenderedPosition) should be Row or Header or "separator".
	if (t.lastRenderedPosition == tw.Row || t.lastRenderedPosition == tw.Header || t.lastRenderedPosition == tw.Position("separator")) &&
		cfg.Settings.Lines.ShowFooterLine.Enabled() {

		t.logger.Debug("streamRenderFooter: Rendering Row/Footer or Header/Footer separator line.")

		// Previous context is the last line rendered before this footer
		prevCells := t.lastRenderedMergeStateToCellContexts(t.lastRenderedLineContent, t.lastRenderedMergeState)

		// Next context is the first line of this footer
		var nextCells map[int]tw.CellContext = nil
		if len(processedFooterLines) > 0 {
			// Need merge states for the footer section.
			// Since footer is processed once and stored, detect merges on its raw input once.
			// This requires access to the *original* raw footer strings passed to Footer().
			// For simplicity now, assume no complex horizontal merges in footer for this separator line context.
			// A better approach: streamStoreFooter should also calculate and store footerMerges.
			// For now, create nextCells without specific merge info for the separator line.
			// Or, call prepareWithMerges on the *stored processed* lines, which might be okay for simple cases.
			// Let's pass nil for sectionMerges to buildStreamCellContexts for this specific Next context.
			// It will result in default (no-merge) states.

			// For now, let's build nextCells manually for the separator line context
			nextCells = make(map[int]tw.CellContext)
			firstFooterLineContent := padLine(processedFooterLines[0], t.streamNumCols)
			// Footer merges should be calculated in streamStoreFooter and stored if needed.
			// For now, assume no merges for this 'Next' context.
			for j := 0; j < t.streamNumCols; j++ {
				nextCells[j] = tw.CellContext{Data: firstFooterLineContent[j], Width: t.streamWidths.Get(j)}
			}
		}

		separatorLevel := tw.LevelFooter // Line before footer section is LevelFooter
		separatorPosition := tw.Footer   // Positioned relative to the footer it precedes
		separatorLocation := tw.LocationMiddle

		f.Line(t.writer, tw.Formatting{
			Row: tw.RowContext{
				Widths:       t.streamWidths,
				ColMaxWidths: tw.CellWidth{PerColumn: t.streamWidths},
				Current:      prevCells, // Context of line above separator
				Previous:     nil,       // No line before Current in this specific context
				Next:         nextCells, // Context of line below separator (first footer line)
				Position:     separatorPosition,
				Location:     separatorLocation,
			},
			Level:            separatorLevel,
			IsSubRow:         false,
			Debug:            t.config.Debug,
			NormalizedWidths: t.streamWidths,
		})
		t.lastRenderedPosition = tw.Position("separator") // Update state
		t.lastRenderedLineContent = nil
		t.lastRenderedMergeState = nil
		t.logger.Debug("streamRenderFooter: Footer separator line rendered.")
	}
	// --- End Render Separator Line ---

	// Detect horizontal merges for the footer section based on its (assumed stored) raw input.
	// This is tricky because streamStoreFooter gets []string, but prepareWithMerges expects [][]string.
	// For simplicity, if complex merges are needed in footer, streamStoreFooter should
	// have received raw data, called prepareWithMerges, and stored those merges.
	// For now, assume no complex horizontal merges in footer or pass nil for sectionMerges.
	// Let's assume footerMerges were calculated and stored as `t.streamFooterMerges map[int]tw.MergeState`
	// by `streamStoreFooter`. For this example, we'll pass nil, meaning no merges.
	var footerMerges map[int]tw.MergeState = nil // Placeholder

	totalFooterLines := len(processedFooterLines)
	for i := 0; i < totalFooterLines; i++ {
		resp := t.buildStreamCellContexts(
			tw.Footer,
			0, // Row index within Footer (always 0)
			i, // Line index
			processedFooterLines,
			footerMerges, // Pass footer-specific merges if calculated and stored
			t.config.Footer,
		)

		// Special Location logic for the *very last line* of the table if this footer line is it.
		// This is complex because bottom border might follow.
		// Let buildStreamCellContexts handle LocationFirst/Middle for now.
		// renderStreamBottomBorder will handle the final LocationEnd for its line.
		// If this footer line is the last content and no bottom border, *it* should be LocationEnd.

		// If this is the last line of the last content block (footer), and no bottom border will be drawn,
		// its Location should be End.
		isLastLineOfTableContent := (i == totalFooterLines-1) &&
			!(cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled())
		if isLastLineOfTableContent {
			resp.location = tw.LocationEnd
			t.logger.Debug("streamRenderFooter: Setting LocationEnd for last footer line as no bottom border will follow.")
		}

		f.Footer(t.writer, [][]string{resp.cellsContent}, tw.Formatting{
			Row: tw.RowContext{
				Widths:       t.streamWidths,
				ColMaxWidths: tw.CellWidth{PerColumn: t.streamWidths},
				Current:      resp.cells,
				Previous:     resp.prevCells,
				Next:         resp.nextCells, // Next is nil if last line of footer block
				Position:     tw.Footer,
				Location:     resp.location,
			},
			Level:            tw.LevelFooter,
			IsSubRow:         (i > 0),
			Debug:            t.config.Debug,
			NormalizedWidths: t.streamWidths,
		})

		t.lastRenderedLineContent = resp.cellsContent
		t.lastRenderedMergeState = make(map[int]tw.MergeState)
		for colIdx, cellCtx := range resp.cells {
			t.lastRenderedMergeState[colIdx] = cellCtx.Merge
		}
		t.lastRenderedPosition = tw.Footer
	}

	t.logger.Debug("streamRenderFooter: Footer content rendering completed.")
	return nil
}

// streamStoreFooter processes the footer content and stores it for later rendering by Close()
// in streaming mode. It ensures stream widths are calculated if not already set.
func (t *Table) streamStoreFooter(footers []string) error {
	t.logger.Debug("streamStoreFooter called with footers: %v", footers)

	// Calculate fixed stream widths if not already set by Start(), Header(), or first Row.
	if t.streamWidths == nil || t.streamWidths.Len() == 0 {
		t.logger.Debug("streamStoreFooter: Stream widths not set, calculating from footer data and config.")
		t.calculateStreamWidths(footers, t.config.Footer) // Pass t.config.Footer (tw.CellConfig)
		if t.streamNumCols == 0 {
			t.logger.Warn("streamStoreFooter: Failed to determine column count from footer data. Footer might not align correctly if other content defines different column count later.")
		} else {
			for i := 0; i < t.streamNumCols; i++ {
				if _, ok := t.streamWidths.OK(i); !ok {
					t.streamWidths.Set(i, 0)
				}
			}
			t.logger.Debug("streamStoreFooter: Determined stream widths from footer: %v", t.streamWidths)
		}
	} else {
		t.logger.Debug("streamStoreFooter: Stream widths already set (%d columns): %v", t.streamNumCols, t.streamWidths)
		if t.streamNumCols > 0 && len(footers) != t.streamNumCols {
			t.logger.Warn("streamStoreFooter: Input footer column count (%d) does not match fixed stream column count (%d). Padding/Truncating input footers.", len(footers), t.streamNumCols)
			if len(footers) < t.streamNumCols {
				paddedFooters := make([]string, t.streamNumCols)
				copy(paddedFooters, footers)
				for i := len(footers); i < t.streamNumCols; i++ {
					paddedFooters[i] = tw.Empty
				}
				footers = paddedFooters
			} else {
				footers = footers[:t.streamNumCols]
			}
		}
	}

	// Only proceed if streamNumCols is determined, otherwise prepareContent will behave unpredictably.
	if t.streamNumCols > 0 {
		t.streamFooterLines = t.prepareContent(footers, t.config.Footer)
		t.logger.Debug("streamStoreFooter: Processed and stored footer lines: %d lines. Content: %v", len(t.streamFooterLines), t.streamFooterLines)
	} else {
		t.logger.Warn("streamStoreFooter: streamNumCols is 0, cannot process/store footer lines meaningfully.")
		t.streamFooterLines = [][]string{} // Ensure it's empty
	}

	return nil
}

// buildStreamCellContexts creates CellContext objects for a given line in streaming mode.
// It determines the Location based on line index within the processed block and overall stream state.
// It constructs Current, Previous, and Next cell contexts using stream widths and lastRenderedState.
// Parameters:
// - position: The section being processed (Header, Row, Footer).
// - rowIdx: The row index within its section (always 0 for Header/Footer, the row number for Row).
// - lineIdx: The line index within the processed lines for this specific row/header/footer block.
// - processedLines: All multi-lines for the current row/header/footer block.
// - sectionMerges: The merge states for the entire section (map[int]tw.MergeState for Header/Footer, map[int]tw.MergeState for the specific row in Row).
// - sectionConfig: The CellConfig for this section (Header, Row, Footer).
// Returns a renderMergeResponse struct containing Current, Previous, Next cells and the determined Location.
func (t *Table) buildStreamCellContexts(
	position tw.Position,
	rowIdx int, // Relevant for Row position
	lineIdx int, // Index within the processedLines slice for this block
	processedLines [][]string,
	sectionMerges map[int]tw.MergeState, // Merges for this section/row block
	sectionConfig tw.CellConfig, // Config for this section (Used for padding/aligns lookup)
) renderMergeResponse {

	resp := renderMergeResponse{
		cells:        make(map[int]tw.CellContext),
		prevCells:    nil,                             // Default to nil
		nextCells:    nil,                             // Default to nil
		cellsContent: make([]string, t.streamNumCols), // Initialize cellsContent
	}

	if t.streamWidths == nil || t.streamWidths.Len() == 0 || t.streamNumCols == 0 {
		t.logger.Warn("buildStreamCellContexts: streamWidths is not set or streamNumCols is 0. Cannot build cell contexts.")
		resp.location = tw.LocationMiddle // Default location
		return resp                       // Return empty contexts
	}

	// Ensure the line exists and is padded to streamNumCols
	currentLineContent := make([]string, t.streamNumCols)
	if lineIdx >= 0 && lineIdx < len(processedLines) {
		currentLineContent = padLine(processedLines[lineIdx], t.streamNumCols)
	} else {
		t.logger.Warn("buildStreamCellContexts: lineIdx %d out of bounds for processedLines (len %d) at position %s, rowIdx %d. Building empty line context.", lineIdx, len(processedLines), position, rowIdx)
		// Populate with empty strings if out of bounds
		for j := range currentLineContent {
			currentLineContent[j] = tw.Empty
		}
	}
	resp.cellsContent = currentLineContent // Store the padded line content

	// Build Current Cells Context
	colAligns := t.buildAligns(sectionConfig)           // Build aligns based on section config
	colPadding := t.buildPadding(sectionConfig.Padding) // Build padding based on section config

	for j := 0; j < t.streamNumCols; j++ {
		cellData := currentLineContent[j]
		finalColWidth := t.streamWidths.Get(j) // Use stream width

		mergeState := tw.MergeState{}
		if sectionMerges != nil {
			if state, ok := sectionMerges[j]; ok {
				mergeState = state
			}
		}
		// Vertical/Hierarchical merges are ignored in streaming.
		// Horizontal merge state comes directly from sectionMerges (detected on raw data).

		resp.cells[j] = tw.CellContext{
			Data:    cellData,
			Align:   colAligns[j],
			Padding: colPadding[j],
			Width:   finalColWidth,
			Merge:   mergeState, // Use the merge state from sectionMerges for this column
		}
	}

	// Determine Previous Cells Context using t.lastRenderedState
	if t.lastRenderedLineContent != nil && t.lastRenderedPosition.Validate() == nil {
		resp.prevCells = make(map[int]tw.CellContext)
		paddedPrevLine := padLine(t.lastRenderedLineContent, t.streamNumCols)
		for j := 0; j < t.streamNumCols; j++ {
			prevMergeState := tw.MergeState{}
			if t.lastRenderedMergeState != nil {
				if state, ok := t.lastRenderedMergeState[j]; ok {
					prevMergeState = state
				}
			}
			resp.prevCells[j] = tw.CellContext{
				Data:  paddedPrevLine[j],
				Width: t.streamWidths.Get(j), // Use stream width for context width
				Merge: prevMergeState,
			}
		}
	}

	// Determine Next Cells Context within the current block
	totalLinesInBlock := len(processedLines)
	if lineIdx < totalLinesInBlock-1 {
		// Next line is within the current processed block (e.g., next line in multi-line header or row)
		resp.nextCells = make(map[int]tw.CellContext)
		nextLineContent := padLine(processedLines[lineIdx+1], t.streamNumCols)
		for j := 0; j < t.streamNumCols; j++ {
			nextMergeState := tw.MergeState{}
			// For the next line within the same block, use the same sectionMerges map
			if sectionMerges != nil {
				if state, ok := sectionMerges[j]; ok {
					nextMergeState = state
				}
			}
			resp.nextCells[j] = tw.CellContext{
				Data:  nextLineContent[j],
				Width: t.streamWidths.Get(j), // Use stream width for context width
				Merge: nextMergeState,
			}
		}
	}
	// If it's the last line of the block, resp.nextCells remains nil by default.
	// Determining the *actual* next line outside the block is handled by the caller (streamRenderHeader, streamAppendRow, Close).

	// Determine Location for this line relative to the *rendering sequence*
	// LocationFirst if: This is the first line of this block (lineIdx == 0) AND the previously rendered element was NOT from the same Position.
	// LocationMiddle otherwise.
	// LocationEnd will be handled for the very last line of the entire stream (in Close or last Append).

	currentRenderingLocation := tw.LocationMiddle // Default

	// Is this the very first content line of this section being rendered?
	isFirstLineOfBlock := (lineIdx == 0)

	// Is the previously rendered element from the same section?
	// If t.lastRenderedPosition is different or invalid, this is the start of a new section block.
	if isFirstLineOfBlock && (t.lastRenderedLineContent == nil || t.lastRenderedPosition != position) {
		currentRenderingLocation = tw.LocationFirst
	} else {
		currentRenderingLocation = tw.LocationMiddle
	}

	resp.location = currentRenderingLocation
	t.logger.Debug("buildStreamCellContexts: Position %s, Row %d, Line %d/%d. Location: %v. Prev Pos: %v. Has Prev: %v.",
		position, rowIdx, lineIdx, totalLinesInBlock, resp.location, t.lastRenderedPosition, t.lastRenderedLineContent != nil)

	return resp
}

// lastRenderedMergeStateToCellContexts converts the stored last rendered line content
// and its merge states into a map of CellContext, suitable for providing
// context (e.g., "Current" or "Previous") to the renderer.
// It uses the fixed streamWidths.
func (t *Table) lastRenderedMergeStateToCellContexts(
	lineContent []string,
	lineMergeStates map[int]tw.MergeState,
) map[int]tw.CellContext {

	cells := make(map[int]tw.CellContext)
	if t.streamWidths == nil || t.streamWidths.Len() == 0 || t.streamNumCols == 0 {
		t.logger.Warn("lastRenderedMergeStateToCellContexts: streamWidths not set or streamNumCols is 0. Returning empty cell contexts.")
		return cells
	}

	// Ensure lineContent is padded to streamNumCols if it's not nil
	var paddedLineContent []string
	if lineContent != nil {
		paddedLineContent = padLine(lineContent, t.streamNumCols)
	} else {
		// If lineContent is nil (e.g. after a separator), create an empty padded line
		paddedLineContent = make([]string, t.streamNumCols)
		for i := range paddedLineContent {
			paddedLineContent[i] = tw.Empty
		}
	}

	for j := 0; j < t.streamNumCols; j++ {
		cellData := paddedLineContent[j]
		colWidth := t.streamWidths.Get(j)
		mergeState := tw.MergeState{} // Default to no merge

		if lineMergeStates != nil {
			if state, ok := lineMergeStates[j]; ok {
				mergeState = state
			}
		}

		// For context purposes (like Previous or Current for a border line),
		// Align and Padding are often less critical than Data, Width, and Merge.
		// We can use default/empty Align and Padding here.
		cells[j] = tw.CellContext{
			Data:    cellData,
			Align:   tw.AlignDefault, // Or tw.AlignNone if preferred for context-only cells
			Padding: tw.Padding{},    // Empty padding
			Width:   colWidth,
			Merge:   mergeState,
		}
	}
	return cells
}
