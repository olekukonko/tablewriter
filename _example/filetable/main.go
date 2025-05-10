package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	folder    = "📁"
	file      = "📄"
	baseDir   = "../"
	indentStr = "    "
)

func main() {
	table := tablewriter.NewTable(os.Stdout, tablewriter.WithTrimSpace(tw.Off))
	table.Header([]string{"Tree", "Size", "Permissions", "Modified"})
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == "." || d.Name() == ".." {
			return nil
		}

		// Calculate relative path depth
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		depth := 0
		if relPath != "." {
			depth = len(strings.Split(relPath, string(filepath.Separator))) - 1
		}

		indent := strings.Repeat(indentStr, depth)

		var name string
		if d.IsDir() {
			name = fmt.Sprintf("%s%s %s", indent, folder, d.Name())
		} else {
			name = fmt.Sprintf("%s%s %s", indent, file, d.Name())
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		table.Append([]string{
			name,
			Size(info.Size()).String(),
			info.Mode().String(),
			Time(info.ModTime()).Format(),
		})

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stdout, "Error: %v\n", err)
		return
	}

	table.Render()
}

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
)

type Size int64

func (s Size) String() string {
	switch {
	case s < KB:
		return fmt.Sprintf("%d B", s)
	case s < MB:
		return fmt.Sprintf("%.2f KB", float64(s)/KB)
	case s < GB:
		return fmt.Sprintf("%.2f MB", float64(s)/MB)
	case s < TB:
		return fmt.Sprintf("%.2f GB", float64(s)/GB)
	default:
		return fmt.Sprintf("%.2f TB", float64(s)/TB)
	}
}

type Time time.Time

func (t Time) Format() string {
	now := time.Now()
	diff := now.Sub(time.Time(t))

	if diff.Seconds() < 60 {
		return "just now"
	} else if diff.Minutes() < 60 {
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	} else if diff.Hours() < 24 {
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	} else if diff.Hours() < 24*7 {
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	} else {
		return time.Time(t).Format("Jan 2, 2006")
	}
}
