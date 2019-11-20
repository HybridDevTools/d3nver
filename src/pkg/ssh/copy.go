package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

type copy struct {
	client       *ssh.Client
	remoteBinary string
}

// Copy a file to a remote location
func (s *SSH) Copy(localFile, remotePath string, mode os.FileMode) (err error) {
	client, err := s.connect()
	if err != nil {
		return
	}
	defer client.Close()

	f, err := os.Open(localFile)
	if err != nil {
		return
	}
	defer f.Close()

	return (&copy{
		client:       client,
		remoteBinary: "scp",
	}).copy(
		f,
		mode,
		remotePath,
	)
}

// https://web.archive.org/web/20170215184048/https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
func (c *copy) copy(file *os.File, mode os.FileMode, destination string) (err error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return
	}

	session, err := c.client.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		return
	}
	defer w.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return
	}

	if err = session.Start(fmt.Sprintf("%s -qt %s", c.remoteBinary, destination)); err != nil {
		return
	}

	go func() {
		err = session.Wait()
		if err != nil {
			if _, ok := err.(*ssh.ExitMissingError); !ok {
				fmt.Println(err)
			}
		}
	}()

	if _, err = fmt.Fprintf(w, "C%#o %d %s\n", mode, fileInfo.Size(), path.Base(destination)); err != nil {
		return
	}
	if _, err = io.Copy(w, file); err != nil {
		return
	}
	if _, err = fmt.Fprint(w, "\x00"); err != nil {
		return
	}

	response, err := c.getResponse(stdout)
	if err != nil {
		return
	}

	if response != "" {
		return errors.New(response)
	}

	return
}

/*
Every message and every finished file data transfer from the provider must be confirmed by the scp process that runs
in a sink mode (= data consumer). The consumer can reply in 3 different messages; binary 0 (OK), 1 (warning)
or 2 (fatal error; will end the connection).
Messages 1 and 2 can be followed by a text message to be printed on the other side, followed by a new line character.
The new line character is mandatory whether the text is empty or not.
*/
func (c *copy) getResponse(reader io.Reader) (message string, err error) {
	buffer := make([]uint8, 1)
	_, err = reader.Read(buffer)
	if err != nil {
		return
	}

	sinkResponse := buffer[0]
	if sinkResponse > 0 {
		bReader := bufio.NewReader(reader)
		return bReader.ReadString('\n')
	}

	return
}
