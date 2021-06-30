package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func serve(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	sess := session.Must(session.NewSession())
	client := s3.New(sess)
	o, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("bias-edge-api-bucket"),
		Key:    aws.String(path),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				err := fmt.Errorf("s3 object %s not found", path)
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&err)
			case s3.ErrCodeInvalidObjectState:
				err := fmt.Errorf("s3 object %s not found", path)
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&err)
			default:
				err := errors.New("internal server error")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&err)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	defer o.Body.Close()
	_, err = io.Copy(w, o.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {

	http.HandleFunc("/", serve)

	addr := ":8090"
	host, port := os.Getenv("HOST"), os.Getenv("PORT")
	if host != "" || port != "" {
		addr = host + ":" + port
	}
	fmt.Printf("Starting on: %s", addr)
	http.ListenAndServe(addr, nil)
}
