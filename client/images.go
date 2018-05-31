package client

import (
	"context"
	"path"

	"github.com/cybozu-go/sabakan"
)

// ImagesIndex get index of images.
func (c *Client) ImagesIndex(ctx context.Context, os string) (sabakan.ImageIndex, *Status) {
	var index sabakan.ImageIndex
	err := c.getJSON(ctx, path.Join("/images", os), nil, &index)
	if err != nil {
		return nil, err
	}
	return index, nil
}
