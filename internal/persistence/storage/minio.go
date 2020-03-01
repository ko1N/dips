package storage

import (
	"fmt"

	"github.com/minio/minio-go"
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
func ConnectMinIO(conf MinIOConfig) (MinIOStorage, error) {
	client, err := minio.New(conf.Endpoint, conf.AccessKey, conf.AccessKeySecret, conf.UseSSL)
	if err != nil {
		return MinIOStorage{}, err
	}
	return MinIOStorage{
		client:   client,
		location: conf.Location,
	}, nil
}

// List - lists files in a remote location
func (s *MinIOStorage) List(bucket string) ([]File, error) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	objectCh := s.client.ListObjects(bucket, "", true, doneCh)

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
func (s *MinIOStorage) CreateBucket(bucket string) error {
	found, err := s.client.BucketExists(bucket)
	if err != nil {
		return err
	}
	if found {
		err = s.DeleteBucket(bucket)
		if err != nil {
			return err
		}
	}
	return s.client.MakeBucket(bucket, s.location)
}

// DeleteBucket - deletes the given storage bucket
func (s *MinIOStorage) DeleteBucket(bucket string) error {
	objectsCh := make(chan string)

	// send objects to the remove channel
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range s.client.ListObjects(bucket, "", true, nil) {
			if object.Err != nil {
				//log.Fatalln(object.Err)
			}
			objectsCh <- object.Key
		}
	}()

	for rErr := range s.client.RemoveObjects(bucket, objectsCh) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

	return s.client.RemoveBucket(bucket)
}

// GetFile - copies a file from the minio storage
func (s *MinIOStorage) GetFile(bucket string, infile string, outpath string) error {
	return s.client.FGetObject(bucket, infile, outpath, minio.GetObjectOptions{})
}

// PutFile - copies a file to the minio storage
func (s *MinIOStorage) PutFile(bucket string, inpath string, outfile string) error {
	_, err := s.client.FPutObject(
		bucket,
		outfile,
		inpath,
		minio.PutObjectOptions{
			//ContentType: "application/octet-stream",
		})
	return err
}
