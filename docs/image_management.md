Boot image management
=====================

In order to distribute network boot images, sabakan provides
an image management system as follows.

* Operators need to upload a boot image only to one sabakan server.

    Rest of sabakan servers will automatically pull the image from the server.

* Sabakan keeps some versions of boot images.

    In case a new image had fatal defects, the change can be rolled back by
    simply removing the new image.

How it works
------------

### Image directory

sabakan saves uploaded images under `/var/lib/sabakan/OS` directory.
`OS` can be an arbitrary identifier such as "coreos".

### Index of images

sabakan keeps an index of available images per OS in etcd.
The structure of the index is a JSON like this:

```json
[
    {
        "id": "1688.5.3",
        "date": "2017-12-02T15:04:05Z",
        "urls": [
            "http://10.1.2.3:10080/api/v1/images/coreos/1688.5.3", 
            "http://10.98.76.54:10080/api/v1/images/coreos/1688.5.3"
        ]
    },
    {
        "id": "1745.4.0",
        "date": "2018-05-29T01:23:45Z",
        "urls": [
            "http://10.1.2.3:10080/api/v1/images/coreos/1745.4.0"
        ]
    }
]
```

`urls` is a list of URLs where the image archive can be downloaded.
Details are described in the next section.

### Watching index

Each sabakan server watches the index in etcd to:
* remove an image if the id of it disappears from the index, and
* pull an image if a new one appears in the index.

Firstly, only one sabakan server in the cluster has a new image.
Other sabakan servers need to pull the image from the first server.

To distribute new images, the first server that receives a new image
adds an index entry having a download URL from the server in `urls`.

When a server finds a new image in the index, it downloads the image through
a URL in `urls`.  After the server pulled the image, it may optionally add
a URL to download the image from itself for load-balancing.

When a server finds an entry disappears in the index, it removes the local
copy of the image.

### Serving requests from iPXE

iPXE downloads a kernel and initial root filesystem image from sabakan.

Sabakan handles these requests from iPXE as follows:

1. Retrieve an image index from etcd.
2. Choose the newest image in the index.
3. Look for the image in the local directory.
4. If the image is found, then return it in the response.
5. If not, choose the next newest image in the index.  Go to 3.
