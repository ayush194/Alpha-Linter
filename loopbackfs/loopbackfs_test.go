package loopbackfs

import (
	"testing"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

func Test(t *testing.T) {
	var ctx context.Context

	lookup_tester := []struct {
		d    *dir
		name string
		node fs.Node
	}{
		{&dir{"testdir"}, "file1.txt", &file{"testdir/file1.txt"}},
		{&dir{"testdir"}, "file2.txt", nil},
		{&dir{"testdir"}, "dir1", &dir{"testdir/dir1"}},
		{&dir{"testdir"}, "somedir", nil},
		{&dir{"testdir/dir1"}, "dir4", &dir{"testdir/dir1/dir4"}},
		{&dir{"testdir/dir1"}, "dir5", nil},
		{&dir{"testdir/dir1"}, "file2.txt", &file{"testdir/dir1/file2.txt"}},
		{&dir{"testdir/dir3"}, "file3.txt", nil},
		{&dir{"testdir/dir3/dir6"}, "somefile.txt", nil},
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
		{&dir{"testdir"}, []fuse.Dirent{{Name: "dir1", Type: fuse.DT_Dir}, {Name: "dir2", Type: fuse.DT_Dir},
			{Name: "dir3", Type: fuse.DT_Dir}, {Name: "file1.txt", Type: fuse.DT_File}}},
		{&dir{"testdir/dir1"}, []fuse.Dirent{{Name: "dir4", Type: fuse.DT_Dir}, {Name: "file2.txt", Type: fuse.DT_File}}},
		{&dir{"testdir/dir1/dir4"}, []fuse.Dirent{{Name: "test.png", Type: fuse.DT_File}}},
		{&dir{"testdir/dir2"}, []fuse.Dirent{{Name: "dir5", Type: fuse.DT_Dir}}},
		{&dir{"testdir/dir2/dir5"}, []fuse.Dirent{{Name: "file1.txt", Type: fuse.DT_Link}}},
		{&dir{"testdir/dir3"}, []fuse.Dirent{{Name: "dir6", Type: fuse.DT_Dir}, {Name: "dir7", Type: fuse.DT_Dir}}},
		{&dir{"testdir/dir3/dir6"}, []fuse.Dirent{{Name: "a.out", Type: fuse.DT_File}}},
		{&dir{"testdir/dir3/dir7"}, []fuse.Dirent{{Name: ".testrc", Type: fuse.DT_File}}},
	}

	for _, testcase := range readdirall_tester {
		res, _ := (testcase.d).ReadDirAll(ctx)
		if len(res) != len(testcase.dirs) {
			t.Errorf("ReadDirAll() returned bad value, expected %v, got %v", testcase.dirs, res)
		}
		for _, res_el := range res {
			no_err := false
			for _, test_el := range testcase.dirs {
				if res_el.Name == test_el.Name && res_el.Type == test_el.Type {
					no_err = true
				}
			}
			if !no_err {
				t.Errorf("ReadDirAll() returned bad value, expected %v, got %v", testcase.dirs, res)
			}
		}
	}

	readlink_tester := []struct {
		f      *file
		req    *fuse.ReadlinkRequest
		target string
	}{
		{&file{"testdir/dir2/dir5/file1.txt"}, &fuse.ReadlinkRequest{}, "../../file1.txt"},
	}

	for _, testcase := range readlink_tester {
		res, _ := (testcase.f).Readlink(ctx, testcase.req)
		if res != testcase.target {
			t.Errorf("Readlink() returned bad value, expected %v, got %v", testcase.target, res)
		}
	}
	/*
		open_file, _ := os.
		write_tester := []struct {
			fh *filehandle
			req *fuse.WriteRequest
		}{
			{}
		}
	*/
}
