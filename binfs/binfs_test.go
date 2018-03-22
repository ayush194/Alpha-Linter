package binfs

import (
	"os"
	"testing"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

func Test(t *testing.T) {
	hellofile := &file{"hellofile", "I just wanted to wish this world hello!", 3, 0444, 1000}
	hellofile2 := &file{"hellofile2", "I just wanted to wish this world hello for a second time!", 4, 0444, 1000}
	hellodir2 := &dir{"hellodir2", 5, os.ModeDir | 0555, nil, nil}
	hellodir := &dir{"hellodir", 2, os.ModeDir | 0555, hellodir2, hellofile2}
	root := &dir{"rootdir", 1, os.ModeDir | 0555, hellodir, hellofile}
	//newfs := filesys{"Ayush", root}
	var ctx context.Context

	lookup_tester := []struct {
		d    *dir
		name string
		node fs.Node
	}{
		{root, "hellofile", hellofile},
		{root, "newfs", nil},
		{hellodir, "hellodir2", hellodir2},
		{hellodir, "hellofile2", hellofile2},
		{hellodir, "newdir", nil},
		{hellodir2, "hellodir2", nil},
		{hellodir2, "", nil},
	}

	for _, testcase := range lookup_tester {
		if res, _ := (testcase.d).Lookup(ctx, testcase.name); testcase.node != res {
			t.Errorf("Lookup() returned bad value, expected %v, got %v", testcase.node, res)
		}
	}

	readdirall_tester := []struct {
		d    *dir
		dirs []fuse.Dirent
	}{
		{root, []fuse.Dirent{{2, fuse.DT_Dir, "hellodir"}, {3, fuse.DT_File, "hellofile"}}},
		{hellodir, []fuse.Dirent{{5, fuse.DT_Dir, "hellodir2"}, {4, fuse.DT_File, "hellofile2"}}},
		{hellodir2, []fuse.Dirent{}},
	}

	for _, testcase := range readdirall_tester {
		res, _ := (testcase.d).ReadDirAll(ctx)
		for idx, dirent := range res {
			if dirent != testcase.dirs[idx] {
				t.Errorf("ReadDirAll() returned bad value, expected %v, got %v", testcase.dirs, res)
			}
		}
	}
}
