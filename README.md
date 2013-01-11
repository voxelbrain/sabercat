`sabercat` is an implementation of [net/http.FileSystem][3] to
serve files directly from [MongoDB's][1] [GridFS][2].

Directory listing has not been implemented due to security
concerns in our particular use case.

[1]: http://www.mongodb.org/
[2]: http://www.mongodb.org/display/DOCS/GridFS
[3]: http://golang.org/pkg/net/http/#FileSystem
---
Version 1.3.0
