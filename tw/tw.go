package tw

// Operation Status Constants
// Used to indicate the success or failure of operations
const (
	Fail    = -1 // Operation failed
	Success = 1  // Operation succeeded
)

const (
	Empty   = ""
	Skip    = ""
	Space   = " "
	NewLine = "\n"
)

// Feature State Constants
// Represents enabled/disabled states for features
const (
	On  State = Success // Feature is enabled
	Off       = Fail    // Feature is disabled
)

type Align string

// Table Alignment Constants
// Defines text alignment options for table content
const (
	AlignNone   Align = "none"   // Center-aligned text
	AlignCenter       = "center" // Center-aligned text
	AlignRight        = "right"  // Right-aligned text
	AlignLeft         = "left"   // Left-aligned text
)

// Position Type and Constants
// Position defines where formatting applies in the table
type Position string

const (
	Header Position = "header" // Table header section
	Row             = "row"    // Table row section
	Footer          = "footer" // Table footer section
)

// Level indicates the vertical position of a line in the table
type Level int

const (
	LevelHeader Level = iota // Topmost line position
	LevelBody                // LevelBody line position
	LevelFooter              // LevelFooter line position
)

type Location string

const (
	LocationFirst  Location = "first"  // Topmost line position
	LocationMiddle Location = "middle" // LevelBody line position
	LocationEnd    Location = "end"    // LevelFooter line position
)

// Text Wrapping Constants
// Defines text wrapping behavior in table cells
const (
	WrapNone     = iota // No wrapping
	WrapNormal          // Standard word wrapping
	WrapTruncate        // Truncate text with ellipsis
	WrapBreak           // Break words to fit
)

// Cell Merge Constants
// Specifies cell merging behavior in tables

const (
	MergeNone         = iota // No merging
	MergeVertical            // Merge cells vertically
	MergeHorizontal          // Merge cells horizontally
	MergeBoth                // Merge both vertically and horizontally
	MergeHierarchical        // Hierarchical merging
)

// Special Character Constants
// Defines special characters used in formatting
const (
	CharEllipsis = "…" // Ellipsis character for truncation
	CharBreak    = "↩" // Break character for wrapping
)
