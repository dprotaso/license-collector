package main

import (
	"archive/tar"
	"fmt"
	"github.com/google/go-containerregistry/pkg/v1"
	"io"
)

type ImageLayers []v1.Layer

func (layers ImageLayers) Walk(onFile func(*tar.Header, io.Reader) error) error {
	for _, layer := range layers {
		ul, err := layer.Uncompressed()
		if err != nil {
			return fmt.Errorf("unable to fetch uncompresed image layer: %v", err)
		}

		tr := tar.NewReader(ul)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break // End of archive
			}
			if err != nil {
				ul.Close()
				return fmt.Errorf("unable to read next file: %v", err)
			}

			if hdr.FileInfo().IsDir() {
				continue
			}

			if err := onFile(hdr, tr); err != nil {
				ul.Close()
				return err
			}
		}

		ul.Close()
	}

	return nil
}
