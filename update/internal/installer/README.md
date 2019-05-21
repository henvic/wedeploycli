# Installer (auto update + run command)
This program runs only once. On its first run, it replaces itself with the newest CLI version and re-executes this new version with the same environment, passing the same arguments, and piping stdin, stderr, and stdout.

## How it works
1. The user installs this program (probably by using a .pkg or .msi installer)
2. The user runs the program
3. Program auto-updates to the newest stable CLI version
4. Program re-executes: forks a new process with its new binary, using the very same environment variable, passing the same arguments, and piping stdin, stderr, stdout

Environment variable `WEDEPLOY_INSTALLER_SKIP_REEXEC` can be used to skip the re-execution step.

## Why
We use [equinox.io](https://equinox.io) to distribute the Liferay Cloud Platform CLI tool (for a few reasons, like its binary deltas). It packages the generated CLI binaries in different formats, such as .rpm, .deb, .msi (Microsoft Installer), and .pkg (macOS installer).

The equinox.io service generates a page with all these different download options for each version we release. However, the .msi and .pkg packages there aren't code signed.

On Unix-like systems, this might be mitigated by recommending our curl based installation process, despite there [are](https://news.ycombinator.com/item?id=12766049) just [too](https://www.idontplaydarts.com/2016/04/detecting-curl-pipe-bash-server-side/) many [reasons](https://sandstorm.io/news/2015-09-24-is-curl-bash-insecure-pgp-verified-install) why this is, overall, a bad idea.

For Windows, we always recommended using the .msi installer given that Windows natively, unfortunately, doesn't come with curl, meaning if we don't code sign the package we end up with unwanted security warnings that might confuse and turn away users.

## Placeholder packages and code signing
While we can't code sign the Equinox distributed packages until they add support for this feature, we are going to take an alternative approach to the ideal (but expensive) case of signing all packages.

Create and serve placeholder packages for macOS and Windows (as described above) so that users can install them on their operating systems without getting security warnings (or even end up completely blocked from installing the software).

The package can try to execute the binary itself, as part of an install script. For this, set the WEDEPLOY_INSTALLER_SKIP_REEXEC environment variable to skip running "lcp" afterward.

Timestamping the signature with a remote server is recommended.

### Windows package
Read the references below to see how it works. There are a few Certificate Authorities selling code sign certificates for Windows, such as [DigiCert](https://www.digicert.com/).

If you can't find the code signing tool, try
`C:\Program Files (x86)\Windows kits\10\App Certification Kit\signtool.exe`.

Currently we use DigiCert, so following the steps described in the document [Signing Code with Microsoft Signcode or SignTool | DigiCert](https://www.digicert.com/code-signing/signcode-signtool-command-line.htm) should be the easier way to code sign a package.

You can verify that an application is signed by right click → File Properties → Digital Signatures. A file can be signed multiple times. Make sure to check the details of the signature to confirm everything is correct.

* [Cryptography Tools](https://docs.microsoft.com/en-us/windows/desktop/seccrypto/cryptography-tools)
* [SignTool](https://docs.microsoft.com/en-us/windows/desktop/seccrypto/signtool)

### macOS package
GateKeeper requires an Apple Developer code signing certificate to avoid a security warning when running a .pkg installer. The only CA recognized by macOS is Apple's own CA.

Tip: verify macOS .pkg installers with the native `installer` program or with [Suspicious Package](https://www.mothersruin.com/software/SuspiciousPackage/).

* Apple's [Code Signing Guide](https://developer.apple.com/library/archive/documentation/Security/Conceptual/CodeSigningGuide/Introduction/Introduction.html)
* [How to sign your Mac OS X App for Gatekeeper](https://successfulsoftware.net/2012/08/30/how-to-sign-your-mac-os-x-app-for-gatekeeper/)
* [panic: About GateKeeper](https://panic.com/blog/about-gatekeeper/)
* [Mac developers: Gatekeeper is a concern, but still gives power users control](https://arstechnica.com/gadgets/2012/02/developers-gatekeeper-a-concern-but-still-gives-power-users-control/)

## Related wedeploy/cli issues
* [Create a signed MSI package for Windows](https://github.com/wedeploy/cli/issues/325)
* [Antivirus found security risk in CLI installer](https://github.com/wedeploy/cli/issues/324)
