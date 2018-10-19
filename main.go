package main

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

var LicenseNames = []string{
	"LICENSE",
	"LICENSE.code",
	"LICENSE.md",
	"LICENSE.txt",
	"COPYING",
	"copyright",
}

var CommonLicensesLocation = []string{
	// Location of common licenses in Debian stretch
	"/usr/share/common-licenses/",
}

func main() {
	log.SetFlags(0)

	image := os.Args[1]

	ref, err := name.ParseReference(image, name.WeakValidation)

	if err != nil {
		log.Fatalf("parsing image reference %q failed: %v", image, err)
	}

	i, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))

	if err != nil {
		log.Fatalf("reading image %q failed: %v", ref, err)
	}

	digest, err := i.Digest()

	if err != nil {
		log.Fatalf("reading digest for %q failed: %v", ref, err)
	}

	layers, err := i.Layers()

	if err != nil {
		log.Fatalf("reading layers for %q failed: %v", ref, err)
	}

	var imageName = ref.Name()

	if tag, ok := ref.(name.Tag); ok {
		imageName = fmt.Sprintf("%s@%s", tag.Name(), digest.String())
	}

	licenseRefs := collectCommonLicenses(layers)
	printLicenses(imageName, layers, licenseRefs)
}

func collectCommonLicenses(layers ImageLayers) map[string]string {
	licenses := make(map[string]string)

	// This code naively assumes if a license links to another
	// it's in the same path
	links := make(map[string]string)

	err := layers.Walk(func(hdr *tar.Header, r io.Reader) error {
		for _, location := range CommonLicensesLocation {
			cleanPath := strings.TrimPrefix(hdr.Name, ".")

			if !strings.HasPrefix(cleanPath, location) {
				continue
			}

			if hdr.Typeflag == tar.TypeLink || hdr.Typeflag == tar.TypeSymlink {
				links[cleanPath] = filepath.Join(location, hdr.Linkname)
			}

			licenseText, err := ioutil.ReadAll(r)

			if err != nil {
				log.Fatalf("unable to read next file: %v", err)
			}

			licenses[cleanPath] = string(licenseText)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Unable to collect common licenses: %v", err)
	}

	for source, target := range links {
		licenses[source] = licenses[target]
	}

	return licenses
}

func printLicenses(imageName string, layers ImageLayers, licenseRefs map[string]string) {
	layers.Walk(func(hdr *tar.Header, r io.Reader) error {
		for _, name := range LicenseNames {
			if strings.HasSuffix(hdr.Name, name) {
				printLicense(imageName, hdr.Name, r, licenseRefs)
			}
		}
		return nil
	})
}

func printLicense(img string, filepath string, reader io.Reader, licenseRefs map[string]string) {
	fmt.Println("===========================================================")
	fmt.Printf("image: %s\n", img)
	fmt.Printf("file:  %s\n", strings.TrimPrefix(filepath, "."))
	fmt.Println("contents:")
	fmt.Println()

	var detectedLicensesRefs []string

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()

		for licensePath := range licenseRefs {
			if strings.Contains(text, licensePath) {
				detectedLicensesRefs = append(detectedLicensesRefs, licensePath)
			}
		}

		printIndented(text)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("unable to read file contents: %v", err)
	}

	for _, ref := range detectedLicensesRefs {
		printIndented("")
		printIndented("contents of common-license file %s", ref)
		printIndented("")
		printCommonLicenseText(licenseRefs[ref])
		printIndented("")
	}

	fmt.Println()
}

func printCommonLicenseText(text string) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		printIndented(scanner.Text())
	}
}

func printIndented(text string, a ...interface{}) {
	fmt.Print("\t")
	fmt.Printf(text, a...)
	fmt.Println()
}
