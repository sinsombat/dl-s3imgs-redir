package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sinsombat/dl-s3imgs-redir/modules"
)

type FormData struct {
	Bucket string                `form:"bucket"`
	File   *multipart.FileHeader `form:"file"`
}

func main() {
	// ENV
	// AWS_ACCESS_KEY_ID
	// AWS_SECRET_ACCESS_KEY
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// server
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/download", download)

	log.Printf("Chi server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Chi server failed to start: %v", err)
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}

	// Parse the data from the request body
	bucket := strings.Trim(r.PostFormValue("bucket"), " ")
	if bucket == "" {
		http.Error(w, "Enter the S3 Bucket name", http.StatusBadRequest)
		return
	}

	suffix := strings.Trim(r.PostFormValue("suffix"), " ")
	if suffix == "" {
		suffix = "-L"
	}

	// Get uploaded file from request
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving uploaded file", http.StatusBadRequest)
		return
	}

	defer file.Close()

	// Read excel data
	xlsx := modules.Xlsx{
		File:    file,
		Handler: header,
	}

	data, err := xlsx.Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Init S3 Client Prepair for Download file
	client := modules.Client{
		Bucket: bucket,
	}
	client.S3Init()

	// get current time
	currentTime := time.Now()

	log.Println("RequestId : ", middleware.GetReqID(r.Context()))
	//modify Re-Structure
	restructureWorker := modules.Restructure{
		Data:    data,
		Client:  client,
		RootDir: currentTime.Format("2006-01-02") + "-" + strings.Join(strings.Split(middleware.GetReqID(r.Context()), "/"), "-"),
		Suffix:  suffix,
	}

	zipFilePath, err := restructureWorker.ModifyDownload()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	//response
	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer zipFile.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+zipFilePath+"\"")
	w.Header().Set("Content-Type", "application/zip")

	os.RemoveAll(restructureWorker.RootDir)
	_, err = io.Copy(w, zipFile)
	os.Remove(zipFilePath)
	if err != nil {
		fmt.Println("Error serving file:", err)
		return
	}
}
