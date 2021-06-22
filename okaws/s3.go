package okaws

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/bitknightwang/okutil/oklog"
	"github.com/bitknightwang/okutil/oksys"
)

func DownloadS3Object(bucket, objectKey, savePath string, isProd bool) error {
	sess, err := CreateAWSSession(isProd)
	if err != nil {
		oklog.Errorf("Failed to create AWS session\n%v", err)
		return err
	}

	downloadFile, err := os.Create(savePath)
	if err != nil {
		oklog.Errorf("Failed to create save file %v\n%v", savePath, err)
		return err
	}

	downloader := s3manager.NewDownloader(sess)
	fileSize, err := downloader.Download(downloadFile, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		oklog.Errorf("Failed to download s3://%v%v -> %v\n%v", bucket, objectKey, savePath, err)
		return err
	}
	oklog.Debugf("file downloaded, %d bytes", fileSize)

	return nil
}

func UploadFileToS3(path, bucket, objectKey string, isProd bool) error {
	if !oksys.IsFile(path) {
		return fmt.Errorf("%v is not a file", path)
	}

	sess, err := CreateAWSSession(isProd)
	if err != nil {
		oklog.Errorf("Failed to create AWS session\n%v", err)
		return err
	}

	s3Service := s3.New(sess)
	uploader := s3manager.NewUploaderWithClient(s3Service)

	// ファイルを開く
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Upload input parameters
	upParams := &s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(DetectFileMimeType(path)),
		Body:        file,
	}

	// Perform upload with options different than the those in the Uploader.
	result, err := uploader.Upload(upParams, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10MB part size
	})
	if err != nil {
		return err
	}

	oklog.Debugf("%v\n", result)
	return nil
}

// ExistsS3Object check file or directory exists in specified s3 bucket
func ExistsInS3(sess *session.Session, bucket, path string) bool {
	if len(bucket) == 0 || len(path) == 0 {
		oklog.Errorf("empty bucket %v or path %v", bucket, path)
		return false
	}

	var objectKey string
	if strings.HasPrefix(path, "/") {
		objectKey = path[1:]
	} else {
		objectKey = path
	}

	s3Service := s3.New(sess)
	result, err := s3Service.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(objectKey),
	})

	if err != nil {
		oklog.Warnf("%v does not exist in %v, error: %v", objectKey, bucket, err)
		return false
	}

	oklog.Debugf("%v check exists in %v result_count:%v", objectKey, bucket, *result.KeyCount)

	return *result.KeyCount > 0
}

func ExistsS3Object(bucket, objectKey string, isProd bool) bool {
	sess, err := CreateAWSSession(isProd)
	if err != nil {
		oklog.Errorf("Failed to create AWS session\n%v", err)
		return false
	}

	return ExistsInS3(sess, bucket, objectKey)
}

// --------------------------------------------------------------------------------------------

// UploadDirToS3 upload a directory to S3 recursively
func UploadDirToS3(dir, bucket, objectKey string, isProd bool) error {
	sess, err := CreateAWSSession(isProd)
	if err != nil {
		oklog.Errorf("Failed to create AWS session\n%v", err)
		return err
	}

	return UploadDirToS3WithSession(sess, dir, bucket, objectKey)
}

func UploadDirToS3WithSession(sess *session.Session, dir, bucket, objectKey string) error {
	if !oksys.IsDir(dir) {
		return fmt.Errorf("%v is not a directory", dir)
	}

	//uploader := s3manager.NewUploader(sess)
	s3Service := s3.New(sess)
	uploader := s3manager.NewUploaderWithClient(s3Service)
	di := NewDirectoryIterator(bucket, objectKey, dir)
	if err := uploader.UploadWithIterator(aws.BackgroundContext(), di); err != nil {
		return fmt.Errorf("failed to upload %v\n%v", dir, err)
	}
	oklog.Debugf("successfully uploaded %v to s3://%v%v", dir, bucket, objectKey)

	return nil
}

// DirectoryIterator represents an iterator of a specified directory
type DirectoryIterator struct {
	baseDir   string
	filePaths []string
	bucket    string
	objectKey string
	next      struct {
		path string
		f    *os.File
	}
	err error
}

// NewDirectoryIterator builds a new DirectoryIterator
func NewDirectoryIterator(bucket, objectKey, dir string) s3manager.BatchUploadIterator {
	var paths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		oklog.Errorf("error walking the path %q: %v\n", dir, err)
		return nil
	}

	return &DirectoryIterator{
		filePaths: paths,
		bucket:    bucket,
		objectKey: objectKey,
		baseDir:   dir,
	}
}

// Next returns whether next file exists or not
func (di *DirectoryIterator) Next() bool {
	if len(di.filePaths) == 0 {
		di.next.f = nil
		return false
	}

	f, err := os.Open(di.filePaths[0])
	di.err = err
	di.next.f = f
	di.next.path = di.filePaths[0]
	di.filePaths = di.filePaths[1:]

	return di.Err() == nil
}

// Err returns error of DirectoryIterator
func (di *DirectoryIterator) Err() error {
	return di.err
}

// UploadObject uploads a file
func (di *DirectoryIterator) UploadObject() s3manager.BatchUploadObject {
	f := di.next.f
	bucketPath := di.next.path[len(di.baseDir):]
	return s3manager.BatchUploadObject{
		Object: &s3manager.UploadInput{
			Bucket:      aws.String(di.bucket),
			Key:         aws.String(di.objectKey + bucketPath),
			ContentType: aws.String(DetectFileMimeType(di.next.path)),
			Body:        f,
		},
		After: func() error {
			return f.Close()
		},
	}
}

func DetectFileMimeType(path string) string {
	// TODO more precisely detect by checking file content
	// contentType := http.DetectContentType(first512BytesOfFile)
	var contentType string
	if strings.HasSuffix(path, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(path, ".gif") {
		contentType = "image/gif"
	} else if strings.HasSuffix(path, ".svg") {
		contentType = "image/svg+xml"
	} else if strings.HasSuffix(path, ".bmp") {
		contentType = "image/bmp"
	} else if strings.HasSuffix(path, ".tif") || strings.HasSuffix(path, ".tiff") {
		contentType = "image/tiff"
	} else if strings.HasSuffix(path, ".htm") || strings.HasSuffix(path, ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(path, ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript"
	} else if strings.HasSuffix(path, ".pdf") {
		contentType = "application/pdf"
	} else if strings.HasSuffix(path, ".ttf") {
		contentType = "font/ttf"
	} else if strings.HasSuffix(path, ".woff") {
		contentType = "font/woff"
	} else if strings.HasSuffix(path, ".otf") {
		contentType = "font/otf"
	} else if strings.HasSuffix(path, ".woff2") {
		contentType = "font/woff2"
	} else if strings.HasSuffix(path, ".eot") {
		contentType = "application/vnd.ms-fontobject"
	} else if strings.HasSuffix(path, ".ico") {
		contentType = "image/vnd.microsoft.icon"
	} else if strings.HasSuffix(path, ".bin") || strings.HasSuffix(path, ".7z") ||
		strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".gz") ||
		strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".tar") ||
		strings.HasSuffix(path, ".bz") || strings.HasSuffix(path, ".bz2") {
		contentType = "application/octet-stream"
	} else {
		contentType = "text/plain"
	}

	return contentType
}
