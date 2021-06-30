package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

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
				fmt.Printf("S3Proxy::ServeRepo::s3.ErrCodeNoSuchKey: %#v", path)
				err := errors.New(fmt.Sprintf("S3 Object %s not found.", path))
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&err)
			case s3.ErrCodeInvalidObjectState:
				fmt.Printf("S3Proxy::ServeRepo::s3.ErrCodeInvalidObjectState: %#v", path)
				err := errors.New(fmt.Sprintf("S3 Object %s not found.", path))
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&err)
			default:
				err := errors.New("Internal Server Error")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&err)
			}
		} else {
			fmt.Printf("S3Proxy::ServeRepo::UnhandledS3Error: %#v", err.Error())
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

	http.ListenAndServe(":8090", nil)
}
