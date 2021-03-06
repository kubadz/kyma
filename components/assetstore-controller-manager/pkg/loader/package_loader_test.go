package loader

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestLoader_Load_Package(t *testing.T) {
	expected := []string{
		"structure/swagger.json",
		"structure/docs/README.md",
	}

	for testName, testCase := range map[string]struct {
		path string
	}{
		"ZipArchive": {
			path: "./testdata/structure.zip",
		},
		"TarGzArchive": {
			path: "./testdata/structure.tar.gz",
		},
		"TarArchive": {
			path: "./testdata/structure.tar.gz",
		},
		"TgzArchive": {
			path: "./testdata/structure.tgz",
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewGomegaWithT(t)

			tmpDir := "../../tmp"
			err := os.MkdirAll(tmpDir, os.ModePerm)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			defer os.RemoveAll(tmpDir)

			loader := &loader{
				temporaryDir:    tmpDir,
				osRemoveAllFunc: os.RemoveAll,
				osCreateFunc:    os.Create,
				httpGetFunc:     getFile(testCase.path),
				ioutilTempDir:   ioutil.TempDir,
			}

			// When
			basePath, files, err := loader.Load(testCase.path, "asset", v1alpha1.AssetPackage)
			defer loader.Clean(basePath)

			// Then
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(files).To(gomega.HaveLen(2))
			g.Expect(files).To(gomega.ConsistOf(expected))
		})
	}
}

func getFile(path string) func(url string) (*http.Response, error) {
	file, err := os.Open(path)
	if err != nil {
		return func(url string) (*http.Response, error) {
			return nil, err
		}
	}

	get := func(url string) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       file,
		}, nil
	}

	return get
}
