package extfs_test

import (
	"bufio"
	"extfs"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"testing"
)

const pathForExtracting = "testExtracted"

type Case struct {
	targetPath        string
	expectedPaths     []string
	expectedFileSizes []int64
}

func getPathsAndSizes(path string) (paths []string, sizes []int64) {
	fsys := os.DirFS(path)
	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		paths = append(paths, p)
		fileInfo, _ := d.Info()
		sizes = append(sizes, fileInfo.Size())
		return nil
	})
	return
}

func createDir(path string) {
	var err error
	err = os.Mkdir(path, 0755)
	if err != nil {
		panic(err)
	}
}

func removeDir(path string) {
	var err error
	err = os.RemoveAll(path)
	if err != nil {
		panic(err)
	}
}

func pruneDir(path string) {
	removeDir(path)
	createDir(path)
}

func TestUnpack(t *testing.T) {
	defer removeDir(pathForExtracting)
	expectedPathsForAverageFWInExt4, expectedSizesForAverageFWInExt4 := parsePtsFile(t, "testImg/FWInExt4.pts")
	cases := []Case{
		{"testImg/emptyFilesExt2.img",
			[]string{".", "lost+found", "test1.txt", "test2.txt"},
			[]int64{4096, 4096, 0, 0}, // respectively
		},
		{"testImg/ext2.img",
			[]string{".", "lost+found", "main", "main/test2.txt", "test1.txt"},
			[]int64{4096, 4096, 4096, 12, 14},
		},
		{
			"testImg/ext3.img",
			[]string{".", "lost+found", "oddloop", "oddloop/amazing", "oddloop/amazing/chu.txt",
				"oddloop/chan.txt", "oddloop/shi.txt", "test.txt"},
			[]int64{4096, 4096, 4096, 4096, 42, 0, 0, 15},
		},
		{
			"testImg/ext4.img",
			[]string{".", "base", "base/slon", "base/slon/test3.txt", "base/test2.txt", "gol",
				"gol/test1.txt", "lost+found"},
			[]int64{4096, 4096, 4096, 12, 11, 4096, 23, 4096},
		},
		{
			targetPath:        "testImg/averageFWInExt4.img",
			expectedPaths:     expectedPathsForAverageFWInExt4,
			expectedFileSizes: expectedSizesForAverageFWInExt4,
		},
	}
	createDir(pathForExtracting)
	for _, c := range cases {
		extfs.Unpack(c.targetPath, pathForExtracting)
		currentPaths, currentSizes := getPathsAndSizes(pathForExtracting)
		if !cmp.Equal(currentPaths, c.expectedPaths) {
			t.Error(cmp.Diff(currentPaths, c.expectedPaths))
		}
		if !cmp.Equal(currentSizes, c.expectedFileSizes) {
			t.Error(cmp.Diff(currentSizes, c.expectedFileSizes))
		}
		pruneDir(pathForExtracting)
	}
}

func parsePtsFile(t *testing.T, path string) (filenames []string, sizes []int64) {
	file, _ := os.Open(path)
	defer file.Close()

	var (
		line           string
		splitLine      []string
		filename, size string
		parsedSize     int64
		err            error
	)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line = scanner.Text()
		splitLine = strings.Split(line, " : ")
		if len(splitLine) == 0 {
			t.Fatal(fmt.Sprintf("file %s has unsupported line", path))
		}
		filename, size = splitLine[0], splitLine[1]
		parsedSize, err = strconv.ParseInt(size, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		filenames = append(filenames, filename)
		if parsedSize != -1 {
			sizes = append(sizes, parsedSize)
		}
	}

	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}

	return
}
