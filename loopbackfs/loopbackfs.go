//Mounts another directory as loopback for testing and benchmarking using Bazil's go-fuse which is a pure Go implementation
package loopbackfs

import (
	"flag"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

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
	fs.Serve(c, &filesys{path})

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}

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

//handle for open file
type filehandle struct {
	open_file *os.File
	path      string
}

/*
//handle for open directory
type dirhandle struct {
	open_dir *os.File
	path     string
}
*/

func (f *filesys) Root() (fs.Node, error) {
	return &dir{f.path}, nil
}

func (d *dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	dir_info, _ := os.Lstat(d.path)
	dir_attr := syscall.Stat_t{}
	syscall.Lstat(d.path, &dir_attr)
	attr.Inode = uint64(dir_attr.Ino)
	attr.Mode = os.FileMode(dir_info.Mode())
	attr.Size = uint64(dir_attr.Size)
	return nil
}

func (d *dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	open_dir, _ := os.Open(d.path)
	defer open_dir.Close()
	dirs, _ := open_dir.Readdir(-1)
	for _, dir_info := range dirs {
		if name == dir_info.Name() {
			if dir_info.IsDir() {
				return &dir{d.path + "/" + name}, nil
			} else {
				return &file{d.path + "/" + name}, nil
			}
		}
	}
	return nil, fuse.ENOENT
}

func (d *dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	open_dir, _ := os.Open(d.path)
	defer open_dir.Close()
	dirs, _ := open_dir.Readdir(-1)
	a := []fuse.Dirent{}
	for _, dir_info := range dirs {
		dir_attr := syscall.Stat_t{}
		syscall.Lstat(d.path+"/"+dir_info.Name(), &dir_attr)
		file_type := []struct {
			check_type os.FileMode
			fuse_type  fuse.DirentType
		}{
			{os.ModeDir, fuse.DT_Dir},
			{os.ModeSymlink, fuse.DT_Link},
			{os.ModeSocket, fuse.DT_Socket},
			{os.ModeCharDevice, fuse.DT_Char},
			{os.ModeNamedPipe, fuse.DT_FIFO},
			{os.FileMode(0xffffffff), fuse.DT_File},
		}
		for _, el := range file_type {
			if dir_info.Mode()&el.check_type != 0 {
				a = append(a, fuse.Dirent{Inode: dir_attr.Ino, Name: dir_info.Name(), Type: el.fuse_type})
				break
			}
		}
	}
	return a, nil
}

/*
func (d *dir) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	dir_info, _ := os.Lstat(d.path)
	//_, err := syscall.Open(d.path, int(req.Flags), uint32(dir_info.Mode()))
	open_dir, err := os.OpenFile(d.path, int(req.Flags), dir_info.Mode())
	//open_dir, err := os.Open(d.path)
	return &dirhandle{open_dir, d.path}, err
}
*/

/*
func (d *dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	return nil
}
*/

/*
func (d *dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	err := os.Mkdir(d.path+"/"+req.Name, req.Mode)
	return d, d, err
}
*/

func (f *file) Attr(ctx context.Context, attr *fuse.Attr) error {
	file_info, _ := os.Lstat(f.path)
	file_attr := syscall.Stat_t{}
	syscall.Lstat(f.path, &file_attr)
	attr.Inode = uint64(file_attr.Ino)
	attr.Mode = os.FileMode(file_info.Mode())
	attr.Size = uint64(file_attr.Size)
	return nil
}

/*
func (f *file) ReadAll(ctx context.Context) ([]byte, error) {
	file, err := os.Open(f.path) // For read access.
	if err != nil {
		return []byte{}, err
	}
	data := make([]byte, 10000000)
	_, err = file.Read(data)
	return data, err
}
*/

func (fh *filehandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	//open_file, _ := os.OpenFile(f.path, os.O_RDONLY, 0666)
	buff := make([]byte, req.Size)
	_, err := fh.open_file.Read(buff)
	//ReadAt already serves the purpose of seek. so no need for another seek call
	//fh.open_file.Seek(int64(n), 1)
	resp.Data = buff
	return err
}

func (fh *filehandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	n, err := fh.open_file.Write(req.Data)
	fh.open_file.Truncate(int64(n))
	resp.Size = int(n)
	return err
}

func (fh *filehandle) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	fh.open_file.Sync()
	return nil
}

func (fh *filehandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	err := fh.open_file.Close()
	return err
}

func (f *file) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	file_info, err := os.Lstat(f.path)
	/*if file_info.Mode()&os.ModeSymlink != 0 {
		target, _ := os.Readlink(f.path)
		path = target
	}*/
	open_file, err := os.OpenFile(f.path, int(req.Flags), file_info.Mode())
	resp.Handle = fuse.HandleID(req.Header.Node)
	//OpenNonSeekable flag makes the OS track the seek in an openfile
	resp.Flags |= fuse.OpenNonSeekable
	return &filehandle{open_file, f.path}, err
}

func (f *file) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	target, err := os.Readlink(f.path)
	return target, err
}

/*
func (f *file) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return nil
}
*/
