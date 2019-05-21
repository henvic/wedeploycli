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
11. git rev-parse HEAD
12. git commit --no-verify --allow-empty --message :msg
13. git remote add :remote-name :remote-address
14. git push :branch --force --no-verify

### Important information
* the status --ignored command is used to retrieve all files ignored by .gitignore files
* configs core.autocrlf and core.safecrlf are unset to avoid warnings regarding mixed [line endings](https://en.wikipedia.org/wiki/Newline) in files
* config user is set because a commit requires these values to be set
* the first call to credential.helper is used to bypass the use of any other credential helper in the project besides the one provided
* :credential-helper is `which lcp` appended with `git-credential-helper`
* :src is a path to a service
* on the commit command, the --no-verify flag bypasses any hooks
* :msg is a git commit message with a JSON specification (see package repodiscovery/tiny).