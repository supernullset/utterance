package main

import(
	"github.com/joho/godotenv"
	"io/ioutil"
	"fmt"
	"net/http"
	"encoding/base64"
	"crypto/rand"
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
	"strings"
	"os"
)

func generateRandomString() string {
	size := 32 // change the length of the generated random string here
	rb := make([]byte,size)
	_, err := rand.Read(rb)
	if err != nil {
		fmt.Println(err)
	}
	rs := base64.URLEncoding.EncodeToString(rb)
	return rs
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling Upload")	
	auth := aws.Auth{
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	}
	locale := aws.USEast
	conn := s3.New(auth, locale)
	bucket := conn.Bucket(os.Getenv("S3_BUCKET"))
	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()
	name := generateRandomString()
	fmt.Printf("Handling %s", header.Header["Content-Type"])

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintln(w, err)
		return 
	}
	mime := header.Header["Content-Type"][0]
	err = bucket.Put(fmt.Sprintf("%s.%s", name, strings.Split(mime, "/")[1]), fileContents, mime, s3.BucketOwnerFull)
	if err != nil {
		fmt.Fprintln(w, "At the S3 Upload")
		fmt.Fprintln(w, err)
		return
	}
	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)
}

func main() {
  err := godotenv.Load()
	if err != nil {
		return
	}
	fmt.Println("Booting server")	
	http.HandleFunc("/", uploadHandler)
	http.ListenAndServe(":8080", nil)
}
