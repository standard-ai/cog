build:	## Build local binary in build/ directory
	@scripts/build.sh

install:	## Install local binary to ~/bin, /usr/local/bin, or $PREFIX
	@scripts/install.sh

help:	## Show this help

release:	## Tag, build, and deploy official release
release:	tag build_release deploy_release

tag:	## Tag a release (auto-increment or use VERSION)
	@scripts/tag.sh

build_release:	## Build binaries in build/ directory for an official release
	rm -rf build/
	@scripts/build.sh darwin amd64
	@scripts/build.sh linux amd64
	@scripts/build.sh linux 386

deploy_release:	## Deploy the official release to github and GCS
	@scripts/deploy_release.sh

deploy_homebrew: ## Deploy homebrew changes
	@scripts/create_homebrew_file.sh

.PHONY: build help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

