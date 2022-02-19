package taskstorage

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// TODO: refactor
type File struct {
	Name        string
	ContentType string
	Size        int64
}

// TODO: use Stat instead of Size() ?
type VirtualFileReader interface {
	io.Reader
	//io.Seeker // optional
	Size() (int64, error)
	io.Closer
}

type VirtualFileWriter interface {
	io.Writer
	//io.Seeker // optional
	io.Closer
}

// Storage -
type Storage interface {
	List(folder string) ([]File, error)

	CreateFolder(folder string) error
	DeleteFolder(folder string) error

	DeleteFile(remotefile string) error

	GetFileReader(fileurl *FileUrl) (VirtualFileReader, error)
	GetFileWriter(fileurl *FileUrl) (VirtualFileWriter, error)

	Close()
}

func ConnectStorage(url *FileUrl) (Storage, error) {
	// connect
	switch url.URL.Scheme {
	case "minio":
		conf := MinIOConfig{
			Endpoint:  url.URL.Host,
			AccessKey: url.URL.User.Username(),
			UseSSL:    false,       // TODO: fix ssl?
			Location:  "us-east-1", // TODO: from path?
		}
		if passwd, ok := url.URL.User.Password(); ok {
			conf.AccessKeySecret = passwd
		}
		return ConnectMinIO(&conf)
	case "smb":
		conf := SmbConfig{
			Server: url.URL.Host,
			User:   url.URL.User.Username(),
		}
		if passwd, ok := url.URL.User.Password(); ok {
			conf.Password = passwd
		}
		return ConnectSmb(&conf)
	default:
		return nil, fmt.Errorf("invalid scheme")
	}
}

type FileUrl struct {
	URL      *url.URL
	Storage  string
	FilePath string
	Dir      string
	FileName string
}

func ParseFileUrl(fileuri string) (*FileUrl, error) {
	url, err := url.Parse(fileuri)
	if err != nil {
		return nil, err
	}

	cleanPath := strings.TrimLeft(path.Clean(url.Path), string(os.PathSeparator))
	splitPath := strings.Split(cleanPath, string(os.PathSeparator))

	var filePath string
	if len(splitPath) == 1 {
		filePath = ""
	} else {
		filePath = strings.Join(splitPath[1:], string(os.PathSeparator))
	}

	dir, fileName := filepath.Split(filePath)

	// TODO: path.Clean() would be better than trimming the suffix
	// but path.Clean() leads to a "." on empty paths
	return &FileUrl{
		URL:      url,
		Storage:  splitPath[0],
		FilePath: filePath,
		Dir:      strings.TrimSuffix(dir, string(os.PathSeparator)),
		FileName: fileName,
	}, nil
}
