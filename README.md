# Go Self-Update
---

A crappy attempt to auto-update running application

- The service pulls binaries from Github Releases checks if `CurrentVersion` < `RemoteVersion`.
- It updates itself if the above is true.

- Do not use this in prod(_not that you would_), I am just experimenting!

## Conclusion:
---

> It kinda works (MacOS could not allow the binary to run - security)

```sh
2025/08/02 21:47:20 INFO: starting service...
2025/08/02 21:47:20 ERROR: unable to start app due to fork/exec /var/folders/1z/rkgl7bkd3572l8bv3y56y6rw0000gn/T/downloads2575145883/selfupdate: permission denied
```

> This was a success

## Meta
---

> cURL requests used

### Check for version
---

```sh
curl --request POST \
  --url http://localhost:8098/version \
  --header 'Content-Type: application/json' \
  --header 'User-Agent: insomnia/2023.5.8' \
  --data '{
	"owner": "TheBeachmaster",
	"repo": "selfupdate"
}'
```

### Upgrade Version
---

__Change URL to match above response__

```sh
curl --request POST \
  --url http://localhost:8098/update \
  --header 'Content-Type: application/json' \
  --header 'User-Agent: insomnia/2023.5.8' \
  --data '{
	"url": "https://github.com/TheBeachMaster/selfupdate/releases/download/2025.8.12/selfupdate_Darwin_arm64.tar.gz"
}'
```
