package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
)

func formRequest() (err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("name", "foo")
	var part io.Writer
	if part, err = writer.CreateFormFile("whatever", "./src/main.rs"); err != nil {
		return
	}
	part.Write([]byte(`fn main(){ println!("hi") }`))

	if part, err = writer.CreateFormFile("foo", "./Cargo.toml"); err != nil {
		return
	}
	part.Write([]byte(`
[package]
name = "foo"
version = "0.0.1"
[dependencies]
embly="0.0.2"
	`))
	if err = writer.Close(); err != nil {
		return
	}

	var req *http.Request
	if req, err = http.NewRequest("POST", "http://api:3000/api/", body); err != nil {
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return err
	}

	fmt.Printf("Got response with status code %d\n", resp.StatusCode)
	var b []byte
	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	fmt.Println(string(b))

	return
}
func main() {
	if err := formRequest(); err != nil {
		log.Fatal(err)
	}
}
