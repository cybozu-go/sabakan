package sabakan

import "time"

const (
	// MaxImages is the maximum number of images that an index can hold.
	MaxImages = 5

	// ImageKernelFilename is a filename appear in TAR archive of an image.
	ImageKernelFilename = "kernel"

	// ImageInitrdFilename is a filename appear in TAR archive of an image.
	ImageInitrdFilename = "initrd.gz"
)

// Image represents a set of image files for iPXE boot.
type Image struct {
	ID   string    `json:"id"`
	Date time.Time `json:"date"`
	URLs []string  `json:"urls"`
}

// ImageIndex is a list of *Image.
type ImageIndex []*Image

// Append appends a new *Image to the index.
//
// If the index has MaxImages images, the oldest image will be discarded.
func (i ImageIndex) Append(img *Image) ImageIndex {
	if len(i) < MaxImages {
		return append(i, img)
	}

	copy(i, i[len(i)-MaxImages+1:len(i)])
	i[MaxImages-1] = img
	return i[0:MaxImages]
}

// Find an image whose ID is id.
//
// If no image can be found, this returns nil.
func (i ImageIndex) Find(id string) *Image {
	for _, img := range i {
		if img.ID == id {
			return img
		}
	}

	return nil
}
