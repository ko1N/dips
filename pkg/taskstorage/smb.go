package taskstorage

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/hirochachacha/go-smb2"
)

// SmbStorage - describes a smb storage
type SmbStorage struct {
	conn    net.Conn
	session *smb2.Session
}

// SmbConfig - config entry describing a storage config
type SmbConfig struct {
	Server   string `json:"server" toml:"server"`
	User     string `json:"user" toml:"user"`
	Password string `json:"password" toml:"password"`
}

// ConnectSmb - opens a connection to smb and returns the connection object
func ConnectSmb(conf *SmbConfig) (*SmbStorage, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:445", conf.Server))
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     conf.User,
			Password: conf.Password,
		},
	}
	session, err := dialer.Dial(conn)
	if err != nil {
		log.Fatalln(err)
		conn.Close()
		return nil, err
	}

	return &SmbStorage{
		conn:    conn,
		session: session,
	}, nil
}

// List - lists files in a remote location
func (self *SmbStorage) List(folder string) ([]File, error) {
	share, dir := parseFilename(folder)

	mount, err := self.session.Mount(share)
	if err != nil {
		return nil, err
	}
	defer mount.Umount()

	fileInfos, err := mount.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []File
	for _, fileInfo := range fileInfos {
		files = append(files, File{
			Name:        fileInfo.Name(),
			ContentType: "application/octet-stream",
			Size:        fileInfo.Size(),
		})
	}
	return files, nil
}

// CreateFolder - creates a new folder
func (self *SmbStorage) CreateFolder(folder string) error {
	share, dir := parseFilename(folder)

	// open mount
	mount, err := self.session.Mount(share)
	if err != nil {
		return err
	}
	defer mount.Umount()

	// make full path
	err = mount.MkdirAll(dir, os.ModeDir)
	if err != nil {
		return err
	}

	return nil
}

// DeleteFolder - deletes the given folder
func (self *SmbStorage) DeleteFolder(folder string) error {
	share, dir := parseFilename(folder)

	// open mount
	mount, err := self.session.Mount(share)
	if err != nil {
		return err
	}
	defer mount.Umount()

	// double check if dir is actually a directory
	stat, err := mount.Stat(dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("cannot delete regular file")
	}

	err = mount.Remove(dir)
	if err != nil {
		return err
	}

	return nil
}

// DeleteFile - deletes a file on the smb storage
func (self *SmbStorage) DeleteFile(remotefile string) error {
	share, remotefilename := parseFilename(remotefile)

	// open mount
	mount, err := self.session.Mount(share)
	if err != nil {
		return err
	}
	defer mount.Umount()

	// double check if dir is actually a file
	stat, err := mount.Stat(remotefilename)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("cannot delete folder")
	}

	err = mount.Remove(remotefilename)
	if err != nil {
		return err
	}

	return nil
}

type SmbFile struct {
	mount *smb2.Share
	file  *smb2.File
}

func (self *SmbFile) Read(p []byte) (n int, err error) {
	return self.file.Read(p)
}

func (self *SmbFile) Write(p []byte) (n int, err error) {
	return self.file.Write(p)
}

func (self *SmbFile) Seek(offset int64, whence int) (int64, error) {
	return self.file.Seek(offset, whence)
}

func (self *SmbFile) Size() (int64, error) {
	info, err := self.file.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (self *SmbFile) Close() error {
	err := self.file.Close()
	if err != nil {
		return err
	}
	return self.mount.Umount()
}

func (self *SmbStorage) GetFileReader(fileurl *FileUrl) (VirtualFileReader, error) {
	// open mount
	mount, err := self.session.Mount(fileurl.Storage)
	if err != nil {
		return nil, err
	}

	// open file
	file, err := mount.Open(fileurl.FilePath)
	if err != nil {
		return nil, err
	}

	return &SmbFile{
		mount: mount,
		file:  file,
	}, nil
}

func (self *SmbStorage) GetFileWriter(fileurl *FileUrl) (VirtualFileWriter, error) {
	// open mount
	mount, err := self.session.Mount(fileurl.Storage)
	if err != nil {
		return nil, err
	}

	// open file
	file, err := mount.Create(fileurl.FilePath)
	if err != nil {
		return nil, err
	}

	return &SmbFile{
		mount: mount,
		file:  file,
	}, nil
}

// Close - closes the samba connection
func (self *SmbStorage) Close() {
	self.session.Logoff()
	self.conn.Close()
}
