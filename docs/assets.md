Asset management
================

In order to help initialization of client servers, sabakan can work
as a file server from which clients can download assets via HTTP.

How it works
------------

### Meta data

Meta data of assets are stored in etcd as follows:

* Each asset has a key like `<prefix>/assets/<NAME>`.
* The value of the key is a JSON like this:

```json
{
    "name": "hoge.tar.gz",
    "id": "15",
    "content-type": "application/tar",
    "date": "2017-12-02T15:04:05Z",
    "sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
    "size": 1002567,
    "urls": [
        "http://10.1.2.3:10080/api/v1/assets/hoge.tar.gz",
        "http://10.98.76.54:10080/api/v1/assets/hoge.tar.gz"
    ],
    "exists": true,
    "options": {
        "version": "3.2.1"
    }
}
```

`id` is a unique numerical ID assigned by sabakan.  Each time a key is
updated, a new `id` will be assigned.

`urls` is a list of URLs where the asset can be downloaded.
Details are described in the next section.

`exists` is only meaningful when this JSON is returned from a REST API.
It becomes `true` if the server has a local copy of the asset.

`options` is optional metadata.  Sabakan just stores and shows these data
as given. Option keys are converted to lowercase implicitly.

### Assets directory

Sabakan saves uploaded assets under `/var/lib/sabakan/assets` directory.

Assets are stored as a file whose name is a numeric ID assigned by sabakan:

```
/var/lib/sabakan/assets/
    - 3
    - 9
    - 15
    - 29
    - ...
```
### Finding and pulling new assets

At first, an asset exists only in the server who received the asset via REST API.
Other sabakan servers need to pull the asset from the first server.

To distribute the new asset, the first server adds meta data to etcd and stores
a download URL from the server in `urls` field.

Each sabakan server watches meta data in etcd, and finds new assets.
When a server finds a new asset, it downloads the asset through a URL in `urls`.
After the server pulled the asset, it may optionally add a URL to download the
asset from itself for load-balancing.

### Removing assets

When a key is removed, the corresponding asset file will also be removed.
When a key is updated, the old asset file will be removed.

### Downloading assets

Clients can download assets from any sabakan server.  If a sabakan server
accepts an asset download request but has no local copy of it, the server
redirects the request to a server in `urls` field.
