Boot image management
=====================

In order to distribute network boot images, sabakan implements
an image management system as follows.

* Operators need to upload a boot image only to one sabakan server.

    Rest of sabakan servers will automatically pull the image from the server.

* Sabakan keeps some versions of boot images.

    If an updated image has fatal defects, the image can be rolled back.
    A REST API to revoke a version is also available.

How it works
------------

### Image directory

sabakan saves uploaded images under `/var/lib/sabakan/<OS>/<VERSION>` directory.

### Index of images

sabakan keeps an index available images in etcd.
The structure of the index look like:

```json
[
    {
        "version": "1688.5.3",
        "servers": ["10.1.2.3", "10.5.67.89"]
    },
    {
        "version": "1745.4.0",
        "servers": ["10.9.8.7"]
    }
]
```

`servers` is a list IP addresses of sabakan servers who has the version locally.
Other sabakan servers access these servers to pull the version.

Each sabakan server watches the index in etcd to:
* remove an image if the version of it disappears from the index, and
* pull an image if a new version appears in the index.
