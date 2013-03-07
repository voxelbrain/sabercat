# Changelog
## 1.4.3
### Bugfixes

* Update CHANGELOG (Oh-so-meta!)

## 1.4.2
### Bugfixes

* Fix a bug where malformed mongodb-URLs would let sabercat panic (no crash)

## 1.4.1
### Bugfixes

* Add regression tests for caching implementation
* Fix a bug where partial responses would be cached
* Add a (hardcoded) upper limit for maximum cache unit size

## 1.4.0
### New features

* Add cluster support (consistency modes)

## 1.3.0
### New features

* Add --strip-slash to strip leading slashes before requesting GridFS

## 1.2.0
### New features

* Add in-memory caching capabilities to standalone version
* Auto-reconnect in connection loss

## 1.1.0
### New features

* Turn `sabercat` into a package rather than a single executable
