# Go Self-Update
---

A crappy attempt to auto-update running application

- The service pulls binaries from Github Releases checks if `CurrentVersion` < `RemoteVersion`.
- It updates itself if the above is true.

- Do not use this in prod(_not that you would_), I am just experimenting!
