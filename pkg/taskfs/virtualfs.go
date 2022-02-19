package taskfs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/ko1N/dips/pkg/taskstorage"
)

type VirtualFS struct {
	// the full path of the root mountpoint of this filesystem
	path string

	filesystem *fuse.Server
	root       *vfsFolder

	inputs  []*taskstorage.FileUrl
	outputs []*taskstorage.FileUrl
}

func (self *VirtualFS) AddInput(input *taskstorage.FileUrl) error {
	for _, in := range self.inputs {
		if in.FilePath == input.FilePath {
			return fmt.Errorf("input file already added\n")
		}
	}

	self.addNode(input, &vfsInputFile{
		input: input,
	})
	self.inputs = append(self.inputs, input)

	return nil
}

func (self *VirtualFS) AddOutput(output *taskstorage.FileUrl) error {
	for _, out := range self.outputs {
		if out.FilePath == output.FilePath {
			return fmt.Errorf("output file already added\n")
		}
	}

	// for outputs we won't add the node unless `Create` is called.
	self.outputs = append(self.outputs, output)

	_, err := self.addFolder(output)
	return err
}

func (self *VirtualFS) RootPath() string {
	return self.path
}

func (self *VirtualFS) ToFullPath(file *taskstorage.FileUrl) (string, error) {
	// path inside of root
	fullpath := path.Join(self.path, file.FilePath)
	containsPath, _ := IsSubPath(self.path, fullpath)
	if !containsPath {
		return "", fmt.Errorf("can't access file outside of sandbox")
	}
	return fullpath, nil
}

// Flush is a no-op in virtual fs
func (self *VirtualFS) Flush() error {
	return nil
}

func (self *VirtualFS) Close() error {
	err := self.filesystem.Unmount()
	if err != nil {
		return err
	}

	if self.path != "" {
		os.RemoveAll(self.path)
	}

	return nil
}

func (self *VirtualFS) addFolder(url *taskstorage.FileUrl) (*fs.Inode, error) {
	p := &self.root.Inode
	dirs := strings.Split(url.Dir, string(os.PathSeparator))

	for idx, component := range dirs {
		if len(component) == 0 {
			continue
		}
		ch := p.GetChild(component)
		if ch == nil {
			ch = p.NewPersistentInode(context.Background(), &vfsFolder{
				vfs:  self,
				path: path.Clean(strings.Join(dirs[:idx+1], string(os.PathSeparator))),
			},
				fs.StableAttr{Mode: fuse.S_IFDIR})
			p.AddChild(component, ch, true)
		}
		p = ch
	}

	return p, nil
}

func (self *VirtualFS) addNode(url *taskstorage.FileUrl, node fs.InodeEmbedder) error {
	p, err := self.addFolder(url)
	if err != nil {
		return err
	}

	ch := p.NewPersistentInode(context.Background(), node, fs.StableAttr{})
	p.AddChild(url.FileName, ch, true)

	return nil
}

type vfsFolder struct {
	fs.Inode

	vfs  *VirtualFS
	path string
}

var _ = (fs.NodeCreater)((*vfsFolder)(nil))

func (self *vfsFolder) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	// check if output is valid
	for _, output := range self.vfs.outputs {
		if output.Dir != self.path { // || output.FileName != name {
			// filename mismatch, exit early
			continue
		}

		ch := self.NewPersistentInode(context.Background(), &vfsOutputFile{
			output: output,
		}, fs.StableAttr{})
		self.AddChild(output.FileName, ch, true)

		// insert in tree
		// file handle can be nil
		// TODO: correct fuseFlags based on fs capabilities?
		return ch, nil, 0 /*fuse.FOPEN_STREAM | fuse.FOPEN_NONSEEKABLE | fuse.FOPEN_DIRECT_IO*/, syscall.F_OK
	}

	// reply with permission denied
	return nil, nil, 0, syscall.EPERM
}

/////////////////////////////////////////////////////

// vfsInputFile - implements a file on the mounted filesystem
type vfsInputFile struct {
	fs.Inode

	input *taskstorage.FileUrl

	mu      sync.Mutex
	storage taskstorage.Storage
	reader  taskstorage.VirtualFileReader
	offset  int64
}

var _ = (fs.NodeGetattrer)((*vfsInputFile)(nil))
var _ = (fs.NodeOpener)((*vfsInputFile)(nil))
var _ = (fs.NodeReader)((*vfsInputFile)(nil))
var _ = (fs.NodeFlusher)((*vfsInputFile)(nil))

// ensureConnection
// ensures the underlying connection and reader exists.
// this function is not mutexed.
func (self *vfsInputFile) ensureConnection() error {
	if self.storage == nil {
		storage, err := taskstorage.ConnectStorage(self.input)
		if err != nil {
			return fmt.Errorf("Unable to connect to storage: %s", err)
		}
		self.storage = storage
	}

	if self.reader == nil {
		reader, err := self.storage.GetFileReader(self.input)
		if err != nil {
			return fmt.Errorf("Failed to get reader from storage: %s", err)
		}
		self.reader = reader
	}

	return nil
}

// Getattr sets the minimum, which is the size. A more full-featured
// FS would also set timestamps and permissions.
func (self *vfsInputFile) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	self.mu.Lock()
	defer self.mu.Unlock()

	err := self.ensureConnection()
	if err != nil {
		return 1 // TODO: proper syscall
	}

	out.Mode = 0755 // TODO:
	out.Nlink = 1
	//out.Mtime = uint64(zf.file.ModTime().Unix())
	//out.Atime = out.Mtime
	//out.Ctime = out.Mtime
	size, _ := self.reader.Size()
	// TODO: error handling
	out.Size = uint64(size)
	const bs = 512
	out.Blksize = bs
	out.Blocks = (out.Size + bs - 1) / bs
	return syscall.F_OK
}

// Open lazily unpacks zip data
func (self *vfsInputFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	self.mu.Lock()
	defer self.mu.Unlock()
	err := self.ensureConnection()
	if err != nil {
		return nil, 0, 1
	}

	// reset offset
	self.offset = 0

	// We don't return a filehandle since we don't really need
	// one.  The file content is immutable, so hint the kernel to
	// cache the data.
	return nil, fuse.FOPEN_KEEP_CACHE, syscall.F_OK
}

// Read simply returns the data that was already unpacked in the Open call
func (self *vfsInputFile) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	self.mu.Lock()
	defer self.mu.Unlock()

	err := self.ensureConnection()
	if err != nil {
		fmt.Printf("Unable to ensure connection: %s\n", err)
		return fuse.ReadResultData([]byte{}), syscall.EIO
	}

	// check if we write sequentially
	// or if seek is supported on the underlying interface
	if self.offset != off {
		if seeker, ok := self.reader.(io.Seeker); ok {
			_, err = seeker.Seek(off, io.SeekStart)
			if err != nil {
				fmt.Printf("Unable to seek: %s\n", err)
				return fuse.ReadResultData([]byte{}), syscall.EIO
			}
		} else {
			fmt.Printf("Unable to do non sequential writes on backend\n")
			return fuse.ReadResultData([]byte{}), syscall.EIO
		}
	}

	var buffer []byte
	fileSize, err := self.reader.Size()
	if err == nil && off+int64(len(dest)) > fileSize {
		// just read to the end of the file
		buffer = make([]byte, fileSize-off)
	} else {
		// fallback, read everything
		buffer = make([]byte, len(dest))
	}

	read, err := self.reader.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Printf("Unable to read data: %s\n", err)
	}
	self.offset += int64(read)

	return fuse.ReadResultData(buffer), syscall.F_OK
}

func (self *vfsInputFile) Flush(ctx context.Context, f fs.FileHandle) syscall.Errno {
	self.mu.Lock()
	defer self.mu.Unlock()

	// close reader and storage and additionally
	// set the internal values to nil.
	// this ensures multiple calls to flush try to close again.
	if self.reader != nil {
		self.reader.Close()
		self.reader = nil
	}

	if self.storage != nil {
		self.storage.Close()
		self.storage = nil
	}

	return syscall.F_OK
}

type vfsOutputFile struct {
	fs.Inode

	output *taskstorage.FileUrl

	mu      sync.Mutex
	storage taskstorage.Storage
	writer  taskstorage.VirtualFileWriter
	offset  int64
}

//var _ = (fs.NodeOpener)((*vfsOutputFile)(nil))
var _ = (fs.NodeWriter)((*vfsOutputFile)(nil))
var _ = (fs.NodeFlusher)((*vfsOutputFile)(nil))

// ensureConnection
// ensures the underlying connection and reader exists.
// this function is not mutexed.
func (self *vfsOutputFile) ensureConnection() error {
	if self.storage == nil {
		storage, err := taskstorage.ConnectStorage(self.output)
		if err != nil {
			return fmt.Errorf("Unable to connect to storage: %s", err)
		}
		self.storage = storage
	}

	if self.writer == nil {
		writer, err := self.storage.GetFileWriter(self.output)
		if err != nil {
			return fmt.Errorf("Failed to get writer for storage: %s", err)
		}
		self.writer = writer
	}

	return nil
}

/*
func (self *vfsOutputFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	self.mu.Lock()
	defer self.mu.Unlock()
	err := self.ensureConnection()
	if err != nil {
		fmt.Printf("Unable to open connection: %s\n", err.Error())
		return nil, 0, 1
	}

	// We don't return a filehandle since we don't really need
	// one.  The file content is immutable, so hint the kernel to
	// cache the data.
	return nil, fuse.FOPEN_STREAM | fuse.FOPEN_NONSEEKABLE | fuse.FOPEN_DIRECT_IO, 0
}
*/

func (self *vfsOutputFile) Write(ctx context.Context, f fs.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {
	self.mu.Lock()
	defer self.mu.Unlock()

	err := self.ensureConnection()
	if err != nil {
		fmt.Printf("Unable to ensure connection: %s\n", err)
		return 0, syscall.EIO
	}

	// check if we write sequentially
	// or if seek is supported on the underlying interface
	if self.offset != off {
		if seeker, ok := self.writer.(io.Seeker); ok {
			_, err = seeker.Seek(off, io.SeekStart)
			if err != nil {
				fmt.Printf("Unable to seek: %s\n", err)
				return 0, syscall.EIO
			}
		} else {
			fmt.Printf("Unable to do non sequential writes on backend\n")
			return 0, syscall.EIO
		}
	}

	written, err := self.writer.Write(data)
	if err != nil {
		fmt.Println(err)
		return 0, syscall.EIO
	}
	self.offset += int64(written)

	return uint32(written), syscall.F_OK
}

func (self *vfsOutputFile) Flush(ctx context.Context, f fs.FileHandle) syscall.Errno {
	self.mu.Lock()
	defer self.mu.Unlock()

	// close reader and storage and additionally
	// set the internal values to nil.
	// this ensures multiple calls to flush try to close again.
	if self.writer != nil {
		self.writer.Close()
		self.writer = nil
	}

	if self.storage != nil {
		self.storage.Close()
		self.storage = nil
	}

	return syscall.F_OK
}

func CreateVirtualFS() (*VirtualFS, error) {
	// create temp folder
	// TODO: make temp storage configurable
	tempFolder := path.Join(".", "temp", uuid.New().String())
	err := os.MkdirAll(tempFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// create fuse filesystem
	opts := &fs.Options{}
	opts.Debug = false

	root := &vfsFolder{}
	server, err := fs.Mount(tempFolder, root, opts)
	if err != nil {
		log.Fatalf("Mount failed: %v\n", err)
		return nil, err
	}
	go func() {
		server.Wait()
	}()

	vfs := &VirtualFS{
		path:       tempFolder,
		filesystem: server,
		root:       root,
	}
	root.vfs = vfs
	return vfs, nil
}
