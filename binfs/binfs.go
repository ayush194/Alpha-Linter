//Creates a binary tree structured filesystem using Bazil's go-fuse which is a pure Go implementation
package binfs

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

/*
func main() {
	flag.Parse()
	mountpoint := flag.Arg(0)
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("ACA-FS"),
		fuse.Subtype("hellofs"),
		fuse.LocalVolume(),
		fuse.VolumeName("Ayush"),
	)
	if err != nil {
		log.Fatal(err)
	}

	//creating my directory structure here
	hellofile := &file{"hellofile", "I just wanted to wish this world hello!", 3, 0666, 1000}
	hellofile2 := &file{"hellofile2", "I just wanted to wish this world hello for a second time!", 4, 0666, 1000}
	hellodir2 := &dir{"hellodir2", 5, os.ModeDir | 0777, nil, nil}
	hellodir := &dir{"hellodir", 2, os.ModeDir | 0777, hellodir2, hellofile2}
	root := &dir{"rootdir", 1, os.ModeDir | 0777, hellodir, hellofile}
	newfs := &filesys{"Ayush", root}

	defer c.Close()
	fs.Serve(c, newfs)

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
	//fuse.Unmount("rootdir")
}
*/

type filesys struct {
	name    string
	rootdir *dir
}

//each dir contains two nodes one of which is a dir and the other a file
type dir struct {
	name     string
	inode    uint64
	mode     os.FileMode
	nextdir  *dir
	nextfile *file
}
type file struct {
	name    string
	content string
	inode   uint64
	mode    os.FileMode
	size    uint64
}

//dir implements node and handle
func (d *dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = d.inode
	attr.Mode = d.mode
	return nil
}

func (d *dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if d.nextdir != nil && name == d.nextdir.name {
		return d.nextdir, nil
	} else if d.nextfile != nil && name == d.nextfile.name {
		return d.nextfile, nil
	}
	return nil, fuse.ENOENT
}

func (d *dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	a := []fuse.Dirent{}
	if d.nextdir != nil {
		a = append(a, fuse.Dirent{Inode: d.nextdir.inode, Name: d.nextdir.name, Type: fuse.DT_Dir})
	}
	if d.nextfile != nil {
		a = append(a, fuse.Dirent{Inode: d.nextfile.inode, Name: d.nextfile.name, Type: fuse.DT_File})
	}
	return a, nil
}

func (d *dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	d.nextdir = &dir{"sub" + req.Name, d.inode + 2, req.Mode, nil, nil}
	return d.nextdir, nil
}

func (f *file) Rename(ctx context.Context, req *fuse.RenameRequest, newFile file) error {
	f.name = req.NewName
	//d.nextdir = newDir
	return nil
}

func (f *file) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.content = string(req.Data)
	resp.Size = 1000
	return nil
}

//file implements node and handle
func (f *file) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = f.inode
	attr.Mode = f.mode
	attr.Size = f.size
	return nil
}

func (f *file) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(f.content), nil
}

//filesys implements FS
func (f *filesys) Root() (fs.Node, error) {
	return f.rootdir, nil
}

/*
func Unmount(dir string) error {
	cmd := exec.Command("umount", dir)
	err := cmd.Run()
	return err
}*/
