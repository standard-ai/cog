build:	## Build local binary in build/ directory
	@scripts/build.sh

install:	## Install local binary to ~/bin, /usr/local/bin, or $PREFIX
	@scripts/install.sh

help:	## Show this help

tag:	## Tag a release (auto-increment or use VERSION)
	@scripts/tag.sh

build_release:	## Build binaries in build/ directory for an official release
	rm -rf build/
	@scripts/build.sh darwin amd64
	@scripts/build.sh linux amd64
	@scripts/build.sh linux 386

deploy_github:	## Deploy the official release to github
	@scripts/deploy_github.sh

deploy_gcs:	## Deploy the official release to github
	@scripts/deploy_gcs.sh

.PHONY: build help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

