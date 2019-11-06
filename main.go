// Compare and contrast files at block level
//
// Craig Hesling <craig@hesling.com>
package main

/*
 * Test case:
 * echo "abc" > test1.txt
 * echo "abC" > test2.txt
 *
 * hd test1.txt
 * hd test2.txt
 *
 * cmpcontrast test1.txt test2.txt
 * The single byte block should show 3/4 bytes matched.
 * Other (large) block sizes should should be 0/1
 *
 * cmpcontrast test1.txt test2.txt --size=3
 * The single byte block should show 2/3 bytes matched.
 * Other (large) block sizes should should be 0/1
 *
 * cmpcontrast test1.txt test2.txt --size=2
 * The single byte block should show 2/2 bytes matched.
 * Other (large) block sizes should should be 1/1
 */

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

var defaultBlockSizes = []int{8 * 1024, 4 * 1024, 2 * 1024, 1024, 512, 256, 1}

func fcompare(f1, f2 *mmap.ReaderAt, bsizes []int, offset int, size int) error {
	var blocks = make(map[int]map[int]bool)

	for _, bsize := range bsizes {
		blocks[int(bsize)] = make(map[int]bool)
	}

	if f1.Len() != f2.Len() {
		fmt.Println("Warning: Files are different sizes")
	}

	/* Scan all bytes of files */
	maxLen := f1.Len()
	if f2.Len() > f1.Len() {
		maxLen = f2.Len()
	}
	if size != -1 {
		if m := offset + size; m < maxLen {
			maxLen = m
		}
	}
	minLen := f1.Len()
	if f2.Len() < f1.Len() {
		minLen = f2.Len()
	}
	for i := offset; i < maxLen; i++ {
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

	/* Analyze results */
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "Block Size\tBlocks-Mismatched\tBlocks-Matched\tBlocks-Total\tPercent Matched\t\n")
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
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%f%%\t\n", bsize, total-matches, matches, total, percent)

	}
	w.Flush()

	return nil
}

func fcompareCmd(cmd *cobra.Command, args []string) error {
	// Fetch requested block sizes
	bsizes, _ := cmd.Flags().GetIntSlice("bsizes")
	offset, _ := cmd.Flags().GetInt("offset")
	size, _ := cmd.Flags().GetInt("size")

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

		if offset != 0 || size != -1 {
			fmt.Printf("# Compare %s vs. %s [off=%d size=%d]\n", f1Name, f2Name, offset, size)
		} else {
			fmt.Printf("# Compare %s vs. %s\n", f1Name, f2Name)
		}

		fcompare(f1, f2, bsizes, offset, size)
	}
	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cmpcontrast <file1> <file2> [files...]",
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

			offset, _ := cmd.Flags().GetInt("offset")
			if offset < 0 {
				return fmt.Errorf("Offset must be non-negative")
			}
			size, _ := cmd.Flags().GetInt("size")
			if size < -1 {
				return fmt.Errorf("Size must be positive or -1")
			}
			return nil
		},
		RunE: fcompareCmd,
	}
	rootCmd.Flags().IntSlice("bsizes", defaultBlockSizes, "List of block sizes to compare against")
	rootCmd.Flags().Int("offset", 0, "Offset to start comparing in byte indices.")
	rootCmd.Flags().Int("size", -1, "Size of region to compare in bytes. A size of -1 means unbounded.")
	rootCmd.Execute()
}
