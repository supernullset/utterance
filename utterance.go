package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

// Parse the upload form template
var form = template.Must(template.New("form").ParseFiles("upload.html"))

// Remove ambiguous characters
var defaultChars = []byte(`abcdefghijkmnpqrstuvwxyz0123456789`)

// RandomString returns random characters from either the given
// variadic bytes, or the default characters
func RandomString(desired int, chars ...byte) string {
	if len(chars) == 0 {
		chars = defaultChars
	}

	possibilities := big.NewInt(int64(len(chars)))
	out := make([]byte, desired)

	for i, _ := range out {
		var index int64
		for {
			r, err := rand.Int(rand.Reader, possibilities)
			if err == nil {
				index = r.Int64()
				break
			}
		}
		out[i] = chars[index]
	}
	return string(out)
}

// server extends the s3 service
type server struct {
	*s3.S3
	bucket string
}

func (srv server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Display the upload form for anything no
	if r.Method != http.MethodPost {
		if err := form.ExecuteTemplate(w, "upload.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	name := RandomString(32)
	out, err := srv.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(srv.bucket),
		Key:         aws.String(name),
		ContentType: aws.String(header.Header.Get("Content-Type")),
		Body:        file,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO anyway to build a bucket URL?

	fmt.Fprintf(w, "name: %s\n", name)
	fmt.Fprintf(w, out.String())
}

// NewServer creates a new server
func NewServer() (srv server) {
	// By default, the ENV credentials provider will look for
	// AWS_ACCESS_KEY(_ID)? and AWS_SECRET(_ACCESS)?_KEY
	credentials := session.New()

	// TODO standard ENV for AWS region? default region?
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}

	srv.S3 = s3.New(credentials, aws.NewConfig().WithRegion(region))
	srv.bucket = os.Getenv("S3_BUCKET")

	// TODO error if the bucket was not set
	return
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("WARNING: %s\n", err)
	}

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}
	host := os.Getenv("HOST")

	fmt.Println("Booting server")
	server := NewServer()
	http.HandleFunc("/", server.uploadHandler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}
