package app

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"fmt"
	"github.com/paraterraform/para/app/index"
	"golang.org/x/net/context"
	"os"
)

// META
const (
	DirRoot  = ""
	FileMeta = ".para"
)

// FUSE

type FS struct {
	index *index.RuntimeIndex
}

func (fs FS) Root() (fs.Node, error) {
	return Dir{path: "", fs: &fs}, nil
}

type Dir struct {
	fs   *FS
	path string
}

func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var result []fuse.Dirent
	if d.path == DirRoot {
		for _, platform := range d.fs.index.ListPlatforms() {
			result = append(result, fuse.Dirent{Name: platform, Type: fuse.DT_Dir})
		}
		result = append(result, fuse.Dirent{Name: FileMeta, Type: fuse.DT_File})
	} else {
		for _, name := range d.fs.index.ListPluginsForPlatform(d.path) {
			result = append(result, fuse.Dirent{Name: name, Type: fuse.DT_File})
		}
	}
	return result, nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if d.path == DirRoot {
		if name == FileMeta {
			return MetaFile{content: fmt.Sprintf("%d\n", os.Getpid())}, nil
		}
		return Dir{path: name, fs: d.fs}, nil
	} else {
		plugin := d.fs.index.LookupPlugin(d.path, name)
		if plugin != nil {
			return File{plugin: plugin, fs: d.fs}, nil
		} else {
			return nil, fuse.ENOENT
		}
	}
}

func (d Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0555
	return nil
}

// File implements both Node and Handle for the hello file.
type File struct {
	plugin *index.Plugin
	fs     *FS
}

func (f File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	err := f.fs.index.OpenPlugin(f.plugin)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	dst := resp.Data[0:req.Size]

	reader, err := f.fs.index.GetReaderAt(f.plugin)
	if err != nil {
		return err
	}

	bytesRead, err := reader.ReadAt(dst, req.Offset)
	resp.Data = dst[:bytesRead]

	return nil
}

func (f File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	err := f.fs.index.ClosePlugin(f.plugin)
	if err != nil {
		return err
	}
	return nil
}

func (f File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0555
	a.Size = f.plugin.Size
	return nil
}

type MetaFile struct {
	content string
}

func (f MetaFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0444
	a.Size = uint64(len(f.content))
	return nil
}

func (f MetaFile) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(f.content), nil
}
