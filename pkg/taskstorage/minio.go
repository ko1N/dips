package taskstorage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage - describes a minio storage
type MinIOStorage struct {
	client   *minio.Client
	location string
}

// MinIOConfig - config entry describing a storage config
type MinIOConfig struct {
	Endpoint        string `json:"endpoint" toml:"endpoint"`
	AccessKey       string `json:"access_key" toml:"access_key"`
	AccessKeySecret string `json:"access_key_secret" toml:"access_key_secret"`
	UseSSL          bool   `json:"use_ssl" toml:"use_ssl"`
	Location        string `json:"location" toml:"location"`
}

// ConnectMinIO - opens a connection to minio and returns the connection object
func ConnectMinIO(conf *MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.AccessKeySecret, ""),
		Secure: conf.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return &MinIOStorage{
		client:   client,
		location: conf.Location,
	}, nil
}

// TODO: deleteme :)
func parseFilename(filename string) (string, string) {
	cleanFilename := strings.TrimLeft(path.Clean(filename), string(os.PathSeparator))
	split := strings.Split(cleanFilename, string(os.PathSeparator))
	if len(split) == 1 {
		return split[0], ""
	} else {
		return split[0], strings.Join(split[1:], string(os.PathSeparator))
	}
}

// List - lists files in a remote location
func (self *MinIOStorage) List(folder string) ([]File, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bucket, prefix := parseFilename(folder)
	if prefix != "" {
		prefix += string(os.PathSeparator)
	}

	objectCh := self.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var files []File
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			continue
		}
		files = append(files, File{
			Name:        object.Key,
			ContentType: object.ContentType,
			Size:        object.Size,
		})
	}
	return files, nil
}

// CreateBucket - creates a new storage bucket
func (self *MinIOStorage) CreateFolder(folder string) error {
	bucket, _ := parseFilename(folder)

	found, err := self.client.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}
	if found {
		err = self.DeleteFolder(bucket)
		if err != nil {
			return err
		}
	}
	return self.client.MakeBucket(
		context.Background(),
		bucket,
		minio.MakeBucketOptions{
			Region:        self.location,
			ObjectLocking: false,
		})
}

// DeleteBucket - deletes the given storage bucket
func (self *MinIOStorage) DeleteFolder(folder string) error {
	bucket, _ := parseFilename(folder)

	objectsCh := make(chan minio.ObjectInfo)

	// send objects to the remove channel
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range self.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
			Prefix:    "",
			Recursive: true,
		}) {
			if object.Err != nil {
				//log.Fatalln(object.Err)
			}
			objectsCh <- object
		}
	}()

	for rErr := range self.client.RemoveObjects(
		context.Background(),
		bucket,
		objectsCh,
		minio.RemoveObjectsOptions{
			GovernanceBypass: false,
		}) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

	return self.client.RemoveBucket(context.Background(), bucket)
}

// DeleteFile - deletes a file on the minio storage
func (self *MinIOStorage) DeleteFile(remotefile string) error {
	bucket, remotefilename := parseFilename(remotefile)
	return self.client.RemoveObject(context.Background(), bucket, remotefilename, minio.RemoveObjectOptions{})
}

type MinIOFileReader struct {
	object *minio.Object
}

func (self *MinIOFileReader) Read(p []byte) (n int, err error) {
	return self.object.Read(p)
}

func (self *MinIOFileReader) Seek(offset int64, whence int) (int64, error) {
	return self.object.Seek(offset, whence)
}

func (self *MinIOFileReader) Size() (int64, error) {
	info, err := self.object.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (self *MinIOFileReader) Close() error {
	return self.object.Close()
}

func (self *MinIOStorage) GetFileReader(fileurl *FileUrl) (VirtualFileReader, error) {
	object, err := self.client.GetObject(context.Background(), fileurl.Storage, fileurl.FilePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &MinIOFileReader{
		object: object,
	}, nil
}

type MinIOFileWriter struct {
	writer *io.PipeWriter
}

func (self *MinIOFileWriter) Write(p []byte) (n int, err error) {
	return self.writer.Write(p)
}

func (self *MinIOFileWriter) Close() error {
	self.writer.CloseWithError(io.EOF)
	return nil
}

func (self *MinIOStorage) GetFileWriter(fileurl *FileUrl) (VirtualFileWriter, error) {
	// create writer pipe
	reader, writer := io.Pipe()

	// TODO: can we expose the error here?
	go func() {
		_, err := self.client.PutObject(context.Background(), fileurl.Storage, fileurl.FilePath, reader, -1, minio.PutObjectOptions{
			//ContentType: "application/octet-stream",
		})
		if err != nil {
			fmt.Printf("PutObject error: %s\n", err)
			//	return nil, err
			return
		}
	}()

	return &MinIOFileWriter{
		writer: writer,
	}, nil
}

// Close - closes the minio connection
func (self *MinIOStorage) Close() {
	// no-op
}
