//Mounts another directory as loopback for testing and benchmarking using Bazil's go-fuse which is a pure Go implementation
package loopbackfs

import (
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

/*
func main() {
	flag.Parse()
	mountpoint, path := flag.Arg(0), flag.Arg(1)

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("ACA-FS"),
		fuse.Subtype("loopbackfs"),
		fuse.LocalVolume(),
		fuse.VolumeName("FUSEFS"),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()
	fs.Serve(c, filesys{path})

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}
*/

//filesystem stores the path to the root directory
type filesys struct {
	path string
}

//each dir contains the path to the actual directory
type dir struct {
	path string
}

//each file contains the path to the actual file
type file struct {
	path string
}

func (f filesys) Root() (fs.Node, error) {
	return dir{f.path}, nil
}

func (d dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	dir_attr := syscall.Stat_t{}
	syscall.Lstat(d.path, &dir_attr)
	attr.Inode = uint64(dir_attr.Ino)
	attr.Mode = os.ModeDir | os.FileMode(dir_attr.Mode)
	attr.Size = uint64(dir_attr.Size)
	return nil
}

func (d dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	open_dir, _ := os.Open(d.path)
	dirs, _ := open_dir.Readdir(6)
	for _, dir_info := range dirs {
		if name == dir_info.Name() {
			if dir_info.IsDir() {
				return dir{d.path + "/" + name}, nil
			} else {
				return file{d.path + "/" + name}, nil
			}
		}
	}
	return nil, fuse.ENOENT
}

func (d dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	open_dir, _ := os.Open(d.path)
	dirs, _ := open_dir.Readdir(6)
	a := []fuse.Dirent{}
	for _, dir_info := range dirs {
		dir_info2 := syscall.Stat_t{}
		syscall.Lstat(d.path+"/"+dir_info.Name(), &dir_info2)
		if dir_info.IsDir() {
			a = append(a, fuse.Dirent{Inode: dir_info2.Ino, Name: dir_info.Name(), Type: fuse.DT_Dir})
		} else {
			a = append(a, fuse.Dirent{Inode: dir_info2.Ino, Name: dir_info.Name(), Type: fuse.DT_File})
		}
	}
	return a, nil
}

func (f file) Attr(ctx context.Context, attr *fuse.Attr) error {
	file_attr := syscall.Stat_t{}
	syscall.Lstat(f.path, &file_attr)
	attr.Inode = uint64(file_attr.Ino)
	attr.Mode = os.FileMode(file_attr.Mode)
	attr.Size = uint64(file_attr.Size)
	return nil
}

func (f file) ReadAll(ctx context.Context) ([]byte, error) {
	file, err := os.Open(f.path) // For read access.
	if err != nil {
		return []byte{}, err
	}
	data := make([]byte, 10000000)
	_, err = file.Read(data)
	return data, err
}
