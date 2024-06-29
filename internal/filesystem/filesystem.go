package filesystem

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

type FS struct {
	fuseutil.NotImplementedFileSystem

	Source DataProvider
}

type FileInfo struct {
	Name  string
	Size  uint64
	IsDir bool
	Inode uint64
}

var (
	rootInode fuseops.InodeID = fuseops.RootInodeID
	nextInode fuseops.InodeID = rootInode + 1
)

func (fs *FS) StatFS(ctx context.Context, op *fuseops.StatFSOp) error {
	return nil
}

func (fs *FS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	log.Printf("GetInodeAttributes")
	info, err := fs.Source.GetChildInfo(uint64(op.Inode), "")
	if err != nil {
		log.Printf("GetInodeAttributes ERROR")
		return err
	}

	op.Attributes = getDefaultAttributes(info)
	return nil
}

func (fs *FS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	log.Printf("LookUpInode")
	info, err := fs.Source.GetChildInfo(uint64(op.Parent), op.Name)
	if err != nil {
		return err
	}

	childInode := fuseops.InodeID(info.Inode)
	//log.Printf("Inodes: %s:%d", path, childInode)
	op.Entry.Child = childInode
	op.Entry.Attributes = getDefaultAttributes(info)
	return nil
}

func (fs *FS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	log.Printf("OpenDir")
	_, err := fs.Source.ListDirectory(uint64(op.Inode))
	return err
}

func (fs *FS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	log.Printf("ReadDir")
	entries, err := fs.Source.ListDirectory(uint64(op.Inode))
	if err != nil {
		return err
	}

	for i := int(op.Offset); i < len(entries); i++ {
		e := entries[i]
		dirent := fuseutil.Dirent{
			Offset: fuseops.DirOffset(i + 1),
			Inode:  fuseops.InodeID(e.Inode),
			Name:   e.Name,
			Type:   getDirEntryType(e.IsDir),
		}
		log.Printf("Inodes: %s:%d", dirent.Name, dirent.Inode)

		n := fuseutil.WriteDirent(op.Dst[op.BytesRead:], dirent)
		if n == 0 {
			break
		}
		op.BytesRead += n
	}

	return nil
}

func (fs *FS) OpenFile(ctx context.Context, op *fuseops.OpenFileOp) error {
	log.Printf("OpenFile")
	return nil
}

func (fs *FS) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
	log.Printf("ReadFile")
	content, err := fs.Source.GetFileContent(uint64(op.Inode))
	if err != nil {
		return err
	}

	if op.Offset > int64(len(content)) {
		return nil
	}

	available := content[op.Offset:]
	n := copy(op.Dst, available)
	op.BytesRead = n

	return nil
}

func getDefaultAttributes(info FileInfo) fuseops.InodeAttributes {
	log.Printf("getDefaultAttributes")
	now := time.Now()
	hardlinks := uint32(1)
	if info.IsDir {
		hardlinks = uint32(2)
	}

	return fuseops.InodeAttributes{
		Size:   info.Size,
		Nlink:  hardlinks, // Number of hard links
		Mode:   getMode(info.IsDir),
		Atime:  now,
		Mtime:  now,
		Ctime:  now,
		Crtime: now,
		Uid:    uint32(os.Getuid()),
		Gid:    uint32(os.Getgid()),
	}
}

func getMode(isDir bool) os.FileMode {
	log.Printf("getMode")
	if isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

func getDirEntryType(isDir bool) fuseutil.DirentType {
	log.Printf("getDirEntryType")
	if isDir {
		return fuseutil.DT_Directory
	}
	return fuseutil.DT_File
}

func Start(mountPoint string, dataProvider DataProvider) {
	if mountPoint == "" {
		log.Fatal("Mount point is required.")
	}

	fs := &FS{
		Source: dataProvider,
	}
	server := fuseutil.NewFileSystemServer(fs)

	mountCfg := &fuse.MountConfig{
		FSName:  "github.com/mstroecker/go-rest-to-fuse",
		Subtype: "go-rest-fuse-v1",
	}

	mfs, err := fuse.Mount(mountPoint, server, mountCfg)
	if err != nil {
		log.Fatalf("Mount failed: %v", err)
	}

	log.Printf("Filesystem mounted. To unmount, use: fusermount -u %s", mountPoint)

	if err = mfs.Join(context.Background()); err != nil {
		log.Fatalf("Join failed: %v", err)
	}
}

type DataProvider interface {
	GetPathForInode(inode uint64) string
	GetChildInfo(parentInode uint64, childName string) (FileInfo, error)
	GetFileContent(inode uint64) ([]byte, error)
	ListDirectory(inode uint64) ([]FileInfo, error)
}
