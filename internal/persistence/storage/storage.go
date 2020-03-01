package storage

// File -
type File struct {
	Name        string
	ContentType string
	Size        int64
}

// Storage -
type Storage interface {
	List(bucket string) ([]File, error)
	CreateBucket(bucket string) error
	DeleteBucket(bucket string) error
	GetFile(bucket string, infile string, outpath string) error
	PutFile(bucket string, inpath string, outfile string) error
}
