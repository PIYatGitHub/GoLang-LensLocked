package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//ImageService is the interfacewe nned for our images
type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) error
	ByGalleryID(galleryID uint) ([]string, error)
}

//imageService is the available API
type imageService struct{}

//NewImageService will init the gallery service and will make it available
func NewImageService() ImageService {
	return &imageService{}
}

func (is *imageService) Create(galleryID uint, r io.ReadCloser, filename string) error {
	defer r.Close()
	path, err := is.mkImagePath(galleryID)
	if err != nil {
		return err
	}

	dst, err := os.Create(path + filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (is *imageService) imagePath(galleryID uint) string {
	return fmt.Sprintf("images/galleries/%v/", galleryID)
}

func (is *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := is.imagePath(galleryID)
	err := os.MkdirAll(galleryPath, 0755) // hackerage...
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}

func (is *imageService) ByGalleryID(galleryID uint) ([]string, error) {
	galleryPath := is.imagePath(galleryID)
	strings, err := filepath.Glob(galleryPath + "*")
	if err != nil {
		return nil, err
	}
	return strings, nil
}
