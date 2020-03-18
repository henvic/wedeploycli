package curlargs

import "strings"

// See https://github.com/curl/curl/blob/master/src/tool_getparam.c
// Last synchronization:
// commit c45360d4633850839bb9c2d77dbf8a8285e9ad49
// date: 11, June 2018.

// curl uses three types of parameters:
// stand-alone but not a boolean, bool (accepts a --no-[name] prefix), and string

// IsCLIStringArgument tells if the command is a cURL CLI string argument
func IsCLIStringArgument(a string) bool {
	for _, alias := range Aliases {
		if a == "-"+alias.Letter || a == "--"+alias.Name {
			return alias.Type == String
		}
	}

	return false
}

// IsCLIArgument tells if the command is a cURL CLI argument
func IsCLIArgument(a string) bool {
	if strings.HasPrefix(a, "-X") {
		return true
	}

	for _, alias := range Aliases {
		if a == "-"+alias.Letter || a == "--"+alias.Name {
			return true
		}
	}

	return false
}

// Argument of the curl CLI tool
type Argument int

const (
	// None means curl accepts no argument parameter
	None Argument = iota
	// Bool means curl accepts either --[name] or --no-[name]
	Bool
	// String means curl accepts a string as parameter for the given argument
	String
)

// LongShort is a description of each option curl accepts
type LongShort struct {
	Letter string
	Name   string
	Type   Argument
}

// Aliases is the list of options curl accepts
var Aliases = []LongShort{
	/* 'letter' strings with more than one character have *no* short option to
	   mention. */
	{"*@", "url", String},
	{"*4", "dns-ipv4-addr", String},
	{"*6", "dns-ipv6-addr", String},
	{"*a", "random-file", String},
	{"*b", "egd-file", String},
	{"*B", "oauth2-bearer", String},
	{"*c", "connect-timeout", String},
	{"*d", "ciphers", String},
	{"*D", "dns-interface", String},
	{"*e", "disable-epsv", Bool},
	{"*f", "disallow-username-in-url", Bool},
	{"*E", "epsv", Bool},
	/* 'epsv' made like this to make --no-epsv and --epsv to work
	   although --disable-epsv is the documented option */
	{"*F", "dns-servers", String},
	{"*g", "trace", String},
	{"*G", "npn", Bool},
	{"*h", "trace-ascii", String},
	{"*H", "alpn", Bool},
	{"*i", "limit-rate", String},
	{"*j", "compressed", Bool},
	{"*J", "tr-encoding", Bool},
	{"*k", "digest", Bool},
	{"*l", "negotiate", Bool},
	{"*m", "ntlm", Bool},
	{"*M", "ntlm-wb", Bool},
	{"*n", "basic", Bool},
	{"*o", "anyauth", Bool},
	{"*p", "wdebug", Bool},
	{"*q", "ftp-create-dirs", Bool},
	{"*r", "create-dirs", Bool},
	{"*s", "max-redirs", String},
	{"*t", "proxy-ntlm", Bool},
	{"*u", "crlf", Bool},
	{"*v", "stderr", String},
	{"*w", "interface", String},
	{"*x", "krb", String},
	{"*x", "krb4", String},
	/* 'krb4' is the previous name */
	{"*X", "haproxy-protocol", Bool},
	{"*y", "max-filesize", String},
	{"*z", "disable-eprt", Bool},
	{"*Z", "eprt", Bool},
	/* 'eprt' made like this to make --no-eprt and --eprt to work
	   although --disable-eprt is the documented option */
	{"*~", "xattr", Bool},
	{"$a", "ftp-ssl", Bool},
	/* 'ftp-ssl' deprecated name since 7.20.0 */
	{"$a", "ssl", Bool},
	/* 'ssl' new option name in 7.20.0, previously this was ftp-ssl */
	{"$b", "ftp-pasv", Bool},
	{"$c", "socks5", String},
	{"$d", "tcp-nodelay", Bool},
	{"$e", "proxy-digest", Bool},
	{"$f", "proxy-basic", Bool},
	{"$g", "retry", String},
	{"$V", "retry-connrefused", Bool},
	{"$h", "retry-delay", String},
	{"$i", "retry-max-time", String},
	{"$k", "proxy-negotiate", Bool},
	{"$m", "ftp-account", String},
	{"$n", "proxy-anyauth", Bool},
	{"$o", "trace-time", Bool},
	{"$p", "ignore-content-length", Bool},
	{"$q", "ftp-skip-pasv-ip", Bool},
	{"$r", "ftp-method", String},
	{"$s", "local-port", String},
	{"$t", "socks4", String},
	{"$T", "socks4a", String},
	{"$u", "ftp-alternative-to-user", String},
	{"$v", "ftp-ssl-reqd", Bool},
	/* 'ftp-ssl-reqd' deprecated name since 7.20.0 */
	{"$v", "ssl-reqd", Bool},
	/* 'ssl-reqd' new in 7.20.0, previously this was ftp-ssl-reqd */
	{"$w", "sessionid", Bool},
	/* 'sessionid' listed as --no-sessionid in the help */
	{"$x", "ftp-ssl-control", Bool},
	{"$y", "ftp-ssl-ccc", Bool},
	{"$j", "ftp-ssl-ccc-mode", String},
	{"$z", "libcurl", String},
	{"$#", "raw", Bool},
	{"$0", "post301", Bool},
	{"$1", "keepalive", Bool},
	/* 'keepalive' listed as --no-keepalive in the help */
	{"$2", "socks5-hostname", String},
	{"$3", "keepalive-time", String},
	{"$4", "post302", Bool},
	{"$5", "noproxy", String},
	{"$7", "socks5-gssapi-nec", Bool},
	{"$8", "proxy1.0", String},
	{"$9", "tftp-blksize", String},
	{"$A", "mail-from", String},
	{"$B", "mail-rcpt", String},
	{"$C", "ftp-pret", Bool},
	{"$D", "proto", String},
	{"$E", "proto-redir", String},
	{"$F", "resolve", String},
	{"$G", "delegation", String},
	{"$H", "mail-auth", String},
	{"$I", "post303", Bool},
	{"$J", "metalink", Bool},
	{"$K", "sasl-ir", Bool},
	{"$L", "test-event", Bool},
	{"$M", "unix-socket", String},
	{"$N", "path-as-is", Bool},
	{"$O", "socks5-gssapi-service", String},
	/* 'socks5-gssapi-service' merged with'proxy-service-name' and
	   deprecated since 7.49.0 */
	{"$O", "proxy-service-name", String},
	{"$P", "service-name", String},
	{"$Q", "proto-default", String},
	{"$R", "expect100-timeout", String},
	{"$S", "tftp-no-options", Bool},
	{"$U", "connect-to", String},
	{"$W", "abstract-unix-socket", String},
	{"$X", "tls-max", String},
	{"$Y", "suppress-connect-headers", Bool},
	{"$Z", "compressed-ssh", Bool},
	{"$~", "happy-eyeballs-timeout-ms", String},
	{"0", "http1.0", None},
	{"01", "http1.1", None},
	{"02", "http2", None},
	{"03", "http2-prior-knowledge", None},
	{"1", "tlsv1", None},
	{"10", "tlsv1.0", None},
	{"11", "tlsv1.1", None},
	{"12", "tlsv1.2", None},
	{"13", "tlsv1.3", None},
	{"1A", "tls13-ciphers", String},
	{"1B", "proxy-tls13-ciphers", String},
	{"2", "sslv2", None},
	{"3", "sslv3", None},
	{"4", "ipv4", None},
	{"6", "ipv6", None},
	{"a", "append", Bool},
	{"A", "user-agent", String},
	{"b", "cookie", String},
	{"B", "use-ascii", Bool},
	{"c", "cookie-jar", String},
	{"C", "continue-at", String},
	{"d", "data", String},
	{"dr", "data-raw", String},
	{"da", "data-ascii", String},
	{"db", "data-binary", String},
	{"de", "data-urlencode", String},
	{"D", "dump-header", String},
	{"e", "referer", String},
	{"E", "cert", String},
	{"Ea", "cacert", String},
	{"Eb", "cert-type", String},
	{"Ec", "key", String},
	{"Ed", "key-type", String},
	{"Ee", "pass", String},
	{"Ef", "engine", String},
	{"Eg", "capath", String},
	{"Eh", "pubkey", String},
	{"Ei", "hostpubmd5", String},
	{"Ej", "crlfile", String},
	{"Ek", "tlsuser", String},
	{"El", "tlspassword", String},
	{"Em", "tlsauthtype", String},
	{"En", "ssl-allow-beast", Bool},
	{"Eo", "login-options", String},
	{"Ep", "pinnedpubkey", String},
	{"EP", "proxy-pinnedpubkey", String},
	{"Eq", "cert-status", Bool},
	{"Er", "false-start", Bool},
	{"Es", "ssl-no-revoke", Bool},
	{"Et", "tcp-fastopen", Bool},
	{"Eu", "proxy-tlsuser", String},
	{"Ev", "proxy-tlspassword", String},
	{"Ew", "proxy-tlsauthtype", String},
	{"Ex", "proxy-cert", String},
	{"Ey", "proxy-cert-type", String},
	{"Ez", "proxy-key", String},
	{"E0", "proxy-key-type", String},
	{"E1", "proxy-pass", String},
	{"E2", "proxy-ciphers", String},
	{"E3", "proxy-crlfile", String},
	{"E4", "proxy-ssl-allow-beast", Bool},
	{"E5", "login-options", String},
	{"E6", "proxy-cacert", String},
	{"E7", "proxy-capath", String},
	{"E8", "proxy-insecure", Bool},
	{"E9", "proxy-tlsv1", None},
	{"EA", "socks5-basic", Bool},
	{"EB", "socks5-gssapi", Bool},
	{"f", "fail", Bool},
	{"fa", "fail-early", Bool},
	{"fb", "styled-output", Bool},
	{"F", "form", String},
	{"Fs", "form-string", String},
	{"g", "globoff", Bool},
	{"G", "get", None},
	{"Ga", "request-target", String},
	{"h", "help", Bool},
	{"H", "header", String},
	{"Hp", "proxy-header", String},
	{"i", "include", Bool},
	{"I", "head", Bool},
	{"j", "junk-session-cookies", Bool},
	{"J", "remote-header-name", Bool},
	{"k", "insecure", Bool},
	{"K", "config", String},
	{"l", "list-only", Bool},
	{"L", "location", Bool},
	{"Lt", "location-trusted", Bool},
	{"m", "max-time", String},
	{"M", "manual", Bool},
	{"n", "netrc", Bool},
	{"no", "netrc-optional", Bool},
	{"ne", "netrc-file", String},
	{"N", "buffer", Bool},
	/* 'buffer' listed as --no-buffer in the help */
	{"o", "output", String},
	{"O", "remote-name", None},
	{"Oa", "remote-name-all", Bool},
	{"p", "proxytunnel", Bool},
	{"P", "ftp-port", String},
	{"q", "disable", Bool},
	{"Q", "quote", String},
	{"r", "range", String},
	{"R", "remote-time", Bool},
	{"s", "silent", Bool},
	{"S", "show-error", Bool},
	{"t", "telnet-option", String},
	{"T", "upload-file", String},
	{"u", "user", String},
	{"U", "proxy-user", String},
	{"v", "verbose", Bool},
	{"V", "version", Bool},
	{"w", "write-out", String},
	{"x", "proxy", String},
	{"xa", "preproxy", String},
	{"X", "request", String},
	{"Y", "speed-limit", String},
	{"y", "speed-time", String},
	{"z", "time-cond", String},
	{"#", "progress-bar", Bool},
	{":", "next", None},
}
