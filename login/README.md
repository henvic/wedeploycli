# Authentication
There are two authentication flows:

* Basic Authentication flow
* OAuth

The first one can be invoked by running

	lcp login --no-browser

It is going to prompt the user for credentials (username and password). After this, it is going to get an OAuth token and continue from there.

The second one (OAuth) is the default and uses the OAuth protocol. More details can be found below.

Another way to authenticate is to pipe thru stdin login:password (similar to using --no-browser) or OAuth token directly.

In this case the CLI connects to API via `POST /login`.

## OAuth based authentication
For more details about how this approach works, see [OAuth 2.0 for Mobile & Desktop Apps](https://developers.google.com/identity/protocols/OAuth2InstalledApp).

In short:

1. The client runs `lcp login`
2. The CLI tool asks the user for permission to open the browser
3. The browser is open on a link that redirects to an HTTP local server running on the CLI tool briefly. The URL looks like this: `https://console.liferay.cloud/login?redirect_uri=http%3A%2F%2Flocalhost%3A65335`
4. The built-in server does all the handshake and sends back the user to an "authorized" [page](https://github.com/wedeploy/cli/blob/7a00f6d2bfeec5e710f6790b24c1a2a442a6465c/loginserver/loginserver.go#L126) on the infrastructure.

Behind the scenes, the CLI tool retrieved the response from the OAuth layer.

After the user logins in Console, the browser is [redirected](https://github.com/wedeploy/cli/blob/7a00f6d2bfeec5e710f6790b24c1a2a442a6465c/loginserver/loginserver.go#L126) with this URL: `http://localhost:65335/#access_token=<access_token>`. The CLI contains a small built-in server that does this [handling](https://github.com/wedeploy/cli/blob/7a00f6d2bfeec5e710f6790b24c1a2a442a6465c/loginserver/loginserver.go#L48) with a blank page with minimal JavaScript to retrieve the token from the fragment part of the redirected URI. This built-in HTTP server runs on an ephemeral port usually for a few seconds or less during the OAuth handshake process.

After the access token is retrieved from the fragment part of the redirected URL, a form is sumbmitted to another endpoint of the local CLI server, called `/authenticate`. This is done, because browser doesn't send the fragment part of the URL to a server.

Once the login process finishes, the user is [redirected](https://github.com/wedeploy/cli/blob/7a00f6d2bfeec5e710f6790b24c1a2a442a6465c/loginserver/loginserver.go#L106) to `https://console.liferay.cloud/cli/login-success` page.


## Known issues on Liferay Cloud platform
The current approach has two major problems that affects the CLI in terms of security and usability.

* The user is not asked for permission on the browser. A grave mistake since the CLI can't know what user is logged in the browser beforehand and this would be a minor security concern, nonetheless. A step should be added on the server-side between steps 3 and 4.
*  The server is reusing the browser's authentication cookie as the OAuth token.

The second point is more painful. It creates a poor user experience and adds security risks.

If the user logs out on the CLI, the browser used to log out too. However, this was removed (not to affect using the browser) meaning that a leaked credential from the CLI means trouble even long after the user types `lcp logout` (especially since there is no feature to revoke existing tokens available).

If the user logs out on the browser, the CLI is logged out.
