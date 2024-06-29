package main

import (
	"flag"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/mstroecker/go-restfuse/internal/filesystem"
	"log"
	"syscall"
)

type TestProvider struct {
}

func (t TestProvider) GetPathForInode(inode uint64) string {
	return "tst"
}

func (t TestProvider) GetFileContent(inode uint64) ([]byte, error) {
	log.Printf("GetFileContent %d", inode)
	switch inode {
	case 100:
		return []byte("Hello__World!"), nil
	case 102:
		return []byte("Content of file1"), nil
	case 103:
		return []byte("Content of file2"), nil
	default:
		return nil, syscall.ENOENT
	}
}

func (t TestProvider) GetChildInfo(parentInode uint64, childName string) (filesystem.FileInfo, error) {
	log.Printf("GetChildInfo %d:%s", parentInode, childName)

	if childName == "" {
		switch parentInode {
		case fuseops.RootInodeID:
			return filesystem.FileInfo{Name: "", Size: 0, IsDir: true, Inode: fuseops.RootInodeID}, nil
		case 100:
			return filesystem.FileInfo{Name: "hello.txt", Size: uint64(len("Hello, World!")), IsDir: false, Inode: 100}, nil
		case 101:
			return filesystem.FileInfo{Name: "subdir", Size: 0, IsDir: true, Inode: 101}, nil
		case 102:
			return filesystem.FileInfo{Name: "file1.txt", Size: uint64(len("Content of file1")), IsDir: false, Inode: 102}, nil
		case 103:
			return filesystem.FileInfo{Name: "file2.txt", Size: uint64(len("Content of file2")), IsDir: false, Inode: 103}, nil
		}
	}

	if parentInode == fuseops.RootInodeID {
		switch childName {
		case "hello.txt":
			return filesystem.FileInfo{Name: "hello.txt", Size: uint64(len("Hello, World!")), IsDir: false, Inode: 100}, nil
		case "subdir":
			return filesystem.FileInfo{Name: "subdir", Size: 0, IsDir: true, Inode: 101}, nil
		}
	} else if parentInode == 101 {
		switch childName {
		case "file1.txt":
			return filesystem.FileInfo{Name: "file1.txt", Size: uint64(len("Content of file1")), IsDir: false, Inode: 102}, nil
		case "file2.txt":
			return filesystem.FileInfo{Name: "file2.txt", Size: uint64(len("Content of file2")), IsDir: false, Inode: 103}, nil
		}
	}

	log.Printf("GetChildInfo ERROR %d:%s", parentInode, childName)
	return filesystem.FileInfo{}, syscall.ENOENT
}

func (t TestProvider) ListDirectory(inode uint64) ([]filesystem.FileInfo, error) {
	log.Printf("listDirectory %d", inode)
	switch inode {
	case fuseops.RootInodeID:
		return []filesystem.FileInfo{
			{Name: "hello.txt", Size: uint64(len("Hello, World!")), IsDir: false, Inode: 100},
			{Name: "subdir", Size: 0, IsDir: true, Inode: 101},
		}, nil
	case 101:
		return []filesystem.FileInfo{
			{Name: "file1.txt", Size: uint64(len("Content of file1")), IsDir: false, Inode: 102},
			{Name: "file2.txt", Size: uint64(len("Content of file2")), IsDir: false, Inode: 103},
		}, nil
	default:
		return nil, syscall.ENOENT
	}
}

func main() {
	mountPoint := flag.String("mount", "", "Mount point for the file system")
	flag.Parse()

	if *mountPoint == "" {
		log.Fatal("Mount point is required. Use --mount flag.")
	}
	log.Printf("Filesystem mounted. To unmount, use: fusermount -u %s", *mountPoint)

	testProvider := TestProvider{}
	// Blocking
	filesystem.Start(*mountPoint, testProvider)
}
