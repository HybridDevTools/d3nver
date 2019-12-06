package natives3

import (
	//"fmt"
	//"io"
	//"net/http"
	//"os"
	//"path/filepath"
	//"strconv"
	//
	//"github.com/cheggaaa/pb/v3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"fmt"
	"os"
)

// NativeS3 storage implementation
type NativeS3 struct{}

// Download a document and place it in destination
func (s *NativeS3) Download(origin, destination string) (destFile string, err error) {

	//fmt.Printf("beh ciao")
	//fmt.Printf(origin) = https://s3-eu-west-1.amazonaws.com/s3.d3nver.io/app/linux/manifest.json
	//fmt.Printf(destination) = /tmp/manifest524602771
	//return

	bucket := "s3.d3nver.io/rbi/stable/virtualbox/"
	item := "box.vdi.bz2"

	file, err := os.Create(destination)
	if err != nil {
		err = fmt.Errorf("unable to open file %q, %v", item, err)
		return
	}

	defer file.Close()

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		err = fmt.Errorf("unable to download item %q, %v", item, err)
		return
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return

	//resp, err := http.Get(origin)
	//if err != nil {
	//	return
	//}
	//defer resp.Body.Close()
	//
	//contentLengthHeader := resp.Header.Get("Content-Length")
	//if contentLengthHeader == "" {
	//	err = fmt.Errorf("cannot determine progress without Content-Length")
	//	return
	//}
	//size, err := strconv.ParseInt(contentLengthHeader, 10, 64)
	//if err != nil {
	//	err = fmt.Errorf("bad Content-Length %q", contentLengthHeader)
	//	return
	//}
	//
	//if resp.StatusCode != 200 {
	//	err = fmt.Errorf("file not found")
	//	return
	//}
	//
	//_, file := filepath.Split(origin)
	//destFile = filepath.Join(destination, file)
	//
	//f, err := os.Create(destFile)
	//if err != nil {
	//	return
	//}
	//defer f.Close()
	//
	//tmpl := `{{ green "Progress:" }} {{counters . | blue}} {{ bar . "[" ("#" | green) ("#" | blue) ("."|white) "]" }} {{percent . | white}} {{speed . }}`
	//bar := pb.ProgressBarTemplate(tmpl).Start64(size)
	//barReader := bar.NewProxyReader(resp.Body)
	//_, err = io.Copy(f, barReader)
	//bar.Finish()
	//
	//return
}
