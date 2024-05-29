package modules

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Directory struct {
	Path string
	Name string
}

type Data struct {
	Source      Directory
	Destination Directory
}

type Restructure struct {
	Data              []Data
	Client            Client
	RootDir           string
	SourceSuffix      string
	DestinationSuffex string
}

func (r *Restructure) ModifyDownload() (zipFilePath string, err error) {
	// create a channel
	c := make(chan int)
	for _, item := range r.Data {
		log.Printf("Row: %+v", item)
		for i := 1; i <= 9; i++ {
			go func(r *Restructure, item Data, i int, c chan int) {
				saveFile(r, item, i)
				c <- i
			}(r, item, i, c)
		}
	}

	for i := 0; i < len(r.Data)*9; i++ {
		// wait for all files to be downloaded
		<-c
	}

	zipFilePath, err = zipFile(r.RootDir)
	return
}

func saveFile(r *Restructure, item Data, i int) (err error) {

	imageNo := strconv.FormatInt(int64(i), 10)
	sourceFileName := item.Source.Path + "/" + item.Source.Name + r.SourceSuffix + imageNo + ".jpg"
	destinationBaseDir := fmt.Sprintf("%s/%s", r.RootDir, item.Destination.Path)
	destinationFileName := fmt.Sprintf("%s/%s%s%d.jpg", destinationBaseDir, item.Destination.Name, r.DestinationSuffex, i)

	ensureBaseDir(destinationBaseDir)

	destinationFile, err := os.Create(destinationFileName)
	if err != nil {
		fmt.Println("failed to create file, ", err)
		return
	}
	defer destinationFile.Close()

	numBytes, err := r.Client.Downloader.Download(destinationFile,
		&s3.GetObjectInput{
			Bucket: aws.String(r.Client.Bucket),
			Key:    aws.String(sourceFileName),
		})

	if err != nil {
		fmt.Println(err, " File : ", sourceFileName)
		os.Remove(destinationFileName)
		return
	}
	fmt.Println("Downloaded", destinationFile.Name(), numBytes, "bytes")
	// Write the downloaded contents to the local file
	return
}

func zipFile(fileDir string) (zipFilePath string, err error) {
	// Create a new zip archive
	zipFilePath = fileDir + ".zip"
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk the folder and add files to the zip archive
	err = filepath.Walk(fileDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a new zip header for the file
		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(filePath, fileDir+string(filepath.Separator))

		// Check if the file is a directory or a regular file
		if fileInfo.IsDir() {
			header.Name += "/"
			header.Method = zip.Store
		} else {
			header.Method = zip.Deflate
		}

		// Create a new zip writer for the file
		if header.Name == (fileDir + string(filepath.Separator)) {
			return err
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// If the file is a regular file, copy its contents to the zip writer
		if !fileInfo.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func ensureBaseDir(baseDir string) error {
	info, err := os.Stat(baseDir)
	if err == nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(baseDir, 0755)
}
