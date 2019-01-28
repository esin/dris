package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"
	"os"
)

type dockerManifestV1 struct {
	SchemaVersion int    `json:"schemaVersion"`
	Name          string `json:"name"`
	Tag           string `json:"tag"`
	Architecture  string `json:"architecture"`
	FsLayers      []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`
	History []struct {
		V1Compatibility time.Time `json:"v1Compatibility"`
	} `json:"history"`
	Signatures []struct {
		Header struct {
			Jwk struct {
				Crv string `json:"crv"`
				Kid string `json:"kid"`
				Kty string `json:"kty"`
				X   string `json:"x"`
				Y   string `json:"y"`
			} `json:"jwk"`
			Alg string `json:"alg"`
		} `json:"header"`
		Signature string `json:"signature"`
		Protected string `json:"protected"`
	} `json:"signatures"`
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	user := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")
	req.SetBasicAuth(user, password)
	return nil
}

func ChunkString(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}

	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}

func ByteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func main() {

  arg := os.Args[len(os.Args)-1] // Last arg

	u, err := url.Parse(arg)
	if err != nil {

	}

	fullImageUrl := fmt.Sprintf("%s://%s/v2/%s", u.Scheme, u.Host, u.Path)
	log.Println("FullImageURL:", fullImageUrl)


	user := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")

	client := &http.Client{
		CheckRedirect: redirectPolicyFunc,
	}

	req, err := http.NewRequest("GET", fullImageUrl+"/manifests/latest", nil)
	req.SetBasicAuth(user, password)

	if err != nil {

	}

	resp, err := client.Do(req)
	if err != nil {

	}

	r, err := ioutil.ReadAll(resp.Body)
	//	log.Println(string(r))

	var manifestParsed dockerManifestV1
	err = json.Unmarshal(r, &manifestParsed)
	if err != nil {

	}

	var imageSize int64
	imageSize = 0
	for _, layer := range manifestParsed.FsLayers {

		log.Println(layer.BlobSum)
		req, err := http.NewRequest("HEAD", fullImageUrl+"/blobs/"+layer.BlobSum, nil)
		req.SetBasicAuth(user, password)

		if err != nil {

		}

		resp, err := client.Do(req)
		if err != nil {

		}

		cmdArgs := []string{"--silent", "--output", "-", "--range", strconv.FormatInt(resp.ContentLength-4, 10) + "-", "-L", "--basic", "--user", user + ":" + password, fullImageUrl + "/blobs/" + layer.BlobSum}
		cmdOut, err := exec.Command("curl", cmdArgs...).Output()
		if err != nil {
			log.Println(err)
		}

		tarSizeHex := []byte{cmdOut[3], cmdOut[2], cmdOut[1], cmdOut[0]}
		tarSize, err := strconv.ParseInt(fmt.Sprintf("0x%x", tarSizeHex), 0, 64)
		if err != nil {
			log.Println(err)
		}

		log.Printf("SIZE: %d, %s", tarSize, ByteCountBinary(tarSize))

		imageSize = imageSize + tarSize

	}

	log.Println("Image size: ", imageSize)
}
