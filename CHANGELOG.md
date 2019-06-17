# Change Log

## Unreleased

### Added

- Automatically download Terraform when it's not present
- Enabled FUSE async read 

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
