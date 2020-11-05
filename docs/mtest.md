Multi-host Test (mtest)
=======================

[mtest](../mtest/) directory contains a test suite to run integration tests.
This suite tests sabactl, sabakan assets, images and ignitions.

Synopsis
--------

[`Makefile`](../mtest/Makefile) setup virtual machine environment and runs mtest.

* `make setup`

    Install mtest required components.

* `make clean`

    Delete generated files in `output/` directory.

* `make placemat`

    Run `placemat` in background by systemd-run to start virtual machines.

* `make stop`

    Stop `placemat`.

* `make test`

    Run mtest on a running `placemat`.
