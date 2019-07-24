# Change Log

## Unreleased

### Fixed

- Output for plugin dir path

## 0.4.1 - 2019-07-19

### Fixed

- IO errors with concurrent open/close operations for same files  

## 0.4.0 - 2019-07-19

### Added

- Auto-unmount stale FUSE mounts over well-known plugin dirs
- PID-based lock next to `plugins` dir so that we can clean screwed FUSE with more confidence

## 0.3.2 - 2019-07-10

### Fixed

- Mount FUSE even if plugin dir is not empty

## 0.3.1 - 2019-06-22

### Fixed

- Failure when extracting plugin archives with directories inside 

## 0.3.0 - 2019-06-18

### Changed

- Para would verify digests of archives before extracting them rather than checking it on the file extracted from it

### Added

- Automatically download Terraform when it's not present
- Automatically download Terragrunt when it's not present
- Enabled FUSE async read 

### Fixed

- FUSE: Plugin dir root no longer returns valid info for anything except valid platform dirs

## 0.2.1 - 2019-06-13

### Fixed

- index cache not being refreshed even if it's older that the refresh interval 
- `failed to read provider plugin ...: input/output error` on Linux due to improper handling of EOF 

## 0.2.0 - 2019-06-13

### Changed

- When extracting archives, look for the 1st file that has prefix `terraform-` in its base name

## 0.1.0 - 2019-06-13

### Added

- Initial implementation
