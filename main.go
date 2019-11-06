// Compare and contrast files at block level
//
// Craig Hesling <craig@hesling.com>
package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/exp/mmap"
)

const (
	BlockMatch    = false
	BlockMismatch = true
)

var defaultBlockSizes = []int{8 * 1024, 4 * 1024, 2 * 1024, 1024, 1}

func fcompare(f1, f2 *mmap.ReaderAt, bsizes []int) error {
	var blocks = make(map[int]map[int]bool)

	for _, bsize := range bsizes {
		blocks[int(bsize)] = make(map[int]bool)
	}

	if f1.Len() != f2.Len() {
		fmt.Println("Warning: Files are different sizes")
	}

	maxLen := f1.Len()
	if f2.Len() > f1.Len() {
		maxLen = f2.Len()
	}
	minLen := f1.Len()
	if f2.Len() < f1.Len() {
		minLen = f2.Len()
	}
	for i := 0; i < maxLen; i++ {
		var match = false
		// If byte is out of bounds for one of the files,
		// assume mismatch
		if i < minLen {
			match = f1.At(i) == f2.At(i)
		}
		for _, bsize := range bsizes {
			if i%bsize == 0 {
				blocks[bsize][i/bsize] = BlockMatch
			}
			if !match {
				blocks[bsize][i/bsize] = BlockMismatch
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "Block Size\tmatched / total\tPercent Matched\t\n")
	for _, bsize := range bsizes {
		var matches, total int
		for _, m := range blocks[bsize] {
			if m == BlockMatch {
				matches++
			}
			total++
		}
		if total == 0 {
			fmt.Println("Warning: We somehow counted 0 block")
			continue
		}
		percent := float64(matches) / float64(total) * 100.0
		fmt.Fprintf(w, "%d\t%d / %d\t%f%%\t\n", bsize, matches, total, percent)

	}
	w.Flush()

	return nil
}

func fcompareCmd(cmd *cobra.Command, args []string) error {
	// Fetch requested block sizes
	bsizes, _ := cmd.Flags().GetIntSlice("bsizes")

	// Open all files
	var fileNames = args
	var files []*mmap.ReaderAt
	for _, fname := range fileNames {
		file, err := mmap.Open(fname)
		if err != nil {
			return err
		}
		files = append(files, file)
		defer file.Close()
	}

	for findex := 0; findex < (len(files) - 1); findex++ {
		if findex > 0 {
			fmt.Println()
		}
		f1 := files[findex]
		f1Name := fileNames[findex]
		f2 := files[findex+1]
		f2Name := fileNames[findex+1]

		sort.Ints(bsizes)

		fmt.Printf("# Compare %s --> %s: %v\n", f1Name, f2Name, bsizes)

		fcompare(f1, f2, bsizes)
	}
	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "filecmpcontrast <file1> <file2> [files...]",
		Short: "Compare one or more files with respect to block size",
		Long:  `Compare multiple files in sequence at different block sizes.`,
		Args:  cobra.MinimumNArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			sizes, err := cmd.Flags().GetIntSlice("bsizes")
			if err != nil {
				return err
			}
			for _, size := range sizes {
				if size < 1 {
					return fmt.Errorf("Block sizes must be positive")
				}
			}
			return nil
		},
		RunE: fcompareCmd,
	}
	rootCmd.Flags().IntSlice("bsizes", defaultBlockSizes, "List of block sizes to compare against")
	rootCmd.Execute()
}
