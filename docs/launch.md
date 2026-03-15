# VirtualPC v1.0 Launch Mode

## Scope
VirtualPC v1.0 is **single-node self-hosted** only.

## Release criteria
Release decision is binary:
- **STATE A: RELEASE READY** when all hard gates pass on real host (`make e2e` and `make soak`).
- **STATE B: RELEASE BLOCKED** when any gate fails.

## Required validation commands
1. `./scripts/build-binaries.sh`
2. `./scripts/build-guest-image.sh`
3. `make e2e`
4. `make soak`
5. `./scripts/build-release.sh`

## Bundle contents
- `bin/virtualpcd`
- `bin/vpc`
- `bin/vpc-agent`
- guest assets (`assets/kernel`, `assets/guest-image.ext4` when present)
- `release-manifest.json`
- `checksums.txt`
- release tarball checksum (`virtualpc-<version>.tar.gz.sha256`)
