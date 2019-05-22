# Transport layer
The transport package contains code responsible for packing a deployment and sending it to Liferay Cloud.

There are two transports, currently:

* [git](https://git-scm.com) (stable)
* [gogit](https://github.com/src-d/go-git) (experimental, not stable)

In theory, *gogit* can replace git allowing the CLI tool to be independent of any external tools. However, gogit is still not [100% compatible](https://github.com/src-d/go-git/blob/master/COMPATIBILITY.md).

Once gogit is more stable, A/B testing can be used to verify if it works better than relying on a git external binary dependency. As of May, 2019 it is better to fail hard with an error message asking the user to install git instead of using a fallback strategy when git is not found.

## Interface

```go
// Transport for the deployment.
type Transport interface {
	Setup(context.Context, transport.Settings) error
	Init() error
	ProcessIgnored() (map[string]struct{}, error)
	Stage(services.ServiceInfoList) error
	Commit(message string) (hash string, err error)
	AddRemote() error
	Push() (groupUID string, err error)
	UploadDuration() time.Duration
	UserAgent() string
}
```

## Commands invoked by transport/git

1. git version
2. git init
3. git status --ignored --untracked-files=all --porcelain -- .
4. git config core.autocrlf false --local
5. git config core.safecrlf false --local
6. git config user.name "Liferay Cloud user" --local
7. git config user.email "user@deployment" --local
8. git config --add credential.helper ""
9. git config --add credential.helper :credential-helper
10. git add :src
11. git commit --no-verify --allow-empty --message :msg
12. git rev-parse HEAD
13. git remote add :remote-name :remote-address
14. git push :branch --force --no-verify

### Deploying to the server
1. The user types `lcp deploy`.
2. CLI finds all `LCP.json` in the current directory and subdirectories.
3. Verify the Git version used.
4. Create a Git repo on a temporary directory, to not clash with any existing Git repo on user's service.
```sh
$ git status --ignored --untracked-files=all --porcelain -- . // Get all ignored files
$ git config core.autocrlf false --local
$ git config core.safecrlf false --local
$ git config user.name "Liferay Cloud user" --local
$ git config user.email "user@deployment" --local
$ git config --add credential.helper ""
$ git config --add credential.helper :credential-helper
```
5. For all services (for every `LCP.json`) the following command is run: `git add :src`
6. We commit to Git using:
```sh
$ git commit --no-verify --allow-empty --message :msg
$ git rev-parse HEAD // Get the SHA1 of the latest commit
```
7. We add the remote. The remote address is [discovered](https://github.com/wedeploy/cli/blob/7a00f6d2bfeec5e710f6790b24c1a2a442a6465c/deployment/transport/git/git.go#L349) via a common convention. The following command is used:
```sh
$ git remote add :remote-name :remote-address
```
8. The code is then pushed to the server using:
```sh
git push :branch --force --no-verify
```

_Note_: There is a [hack](https://github.com/wedeploy/cli/blob/7a00f6d2bfeec5e710f6790b24c1a2a442a6465c/deployment/transport/git/pushhack.go#L37) for Windows and some [versions](https://github.com/wedeploy/cli/blob/7a00f6d2bfeec5e710f6790b24c1a2a442a6465c/deployment/transport/git/pushhack.go#L24) of Git on pushing the code.

About 4). The command used was:
```sh
$ lcp deploy --copy-pkg ./
```

```sh
➜  test-hosting-dxpcloud git:(master) ✗ l
total 32
-rw-r--r--@  1 iliyan  staff   6.0K May 31  2018 .DS_Store
drwxr-xr-x  13 iliyan  staff   416B May 22 16:03 .git
-rw-r--r--   1 iliyan  staff    85B May 10 15:13 LCP.json
-rw-r--r--   1 iliyan  staff    95B Jan 29 09:42 index.html
drwx------   4 iliyan  staff   128B May 22 16:01 myprj3-2019-05-22-16-01-28+0200
➜  test-hosting-dxpcloud git:(master) ✗ cd myprj3-2019-05-22-16-01-28+0200 
➜  myprj3-2019-05-22-16-01-28+0200 git:(master) l
total 0
drwxr-xr-x  13 iliyan  staff   416B May 22 16:03 .git
drwxr-xr-x   4 iliyan  staff   128B May 22 16:01 test-hosting-dxpcloud
➜  myprj3-2019-05-22-16-01-28+0200 git:(master) lcp deploy --copy-pkg ./          
```


### Important information
* the status --ignored command is used to retrieve all files ignored by .gitignore files
* configs core.autocrlf and core.safecrlf are unset to avoid warnings regarding mixed [line endings](https://en.wikipedia.org/wiki/Newline) in files
* config user is set because a commit requires these values to be set
* the first call to credential.helper is used to bypass the use of any other credential helper in the project besides the one provided
* :credential-helper is `which lcp` appended with `git-credential-helper`
* :src is a path to a service
* on the commit command, the --no-verify flag bypasses any hooks
* :msg is a git commit message with a JSON specification (see package repodiscovery/tiny).
