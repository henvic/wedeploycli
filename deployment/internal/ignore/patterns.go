package ignore

// Patterns to ignore (filepath.Match based)
// Mostly based on what is found on github.com/github/gitignore, but more limited.
var Patterns = []string{
	// obviously we don't want to copy .git!
	".git",

	// From Global/macOS.gitignore: General
	".DS_Store",
	".AppleDouble",
	".LSOverride",

	// From Global/macOS.gitignore: Directories potentially created on remote AFP share
	".AppleDB",
	".AppleDesktop",
	"Network Trash Folder",
	"Temporary Items",
	".apdisk",

	// From Global/Xcode.gitignore
	"xcuserdata",

	// From Global/Windows.gitignore: Windows thumbnail cache files
	"Thumbs.db",
	"ehthumbs.db",
	"ehthumbs_vista.db",

	// From Global/Windows.gitignore: Dump file
	"*.stackdump",
	"Desktop.ini",
	"desktop.ini",

	// From Global/VisualStudioCode.gitignore
	".vscode",

	// From Global/Vim.gitignore: Swap
	"[._]*.s[a-v][a-z]",
	"[._]*.sw[a-p]",
	"[._]s[a-v][a-z]",
	"[._]sw[a-p]",

	// From Global/Vim.gitignore: Session
	"Session.vim",

	// From Global/Vim.gitignore: Temporary
	".netrwhist",

	// From Global/Vagrant.gitignore
	".vagrant",

	// From Global/SVN.gitignore
	".svn",

	// From Global/Mercurial.gitignore
	".hg",

	// From Global/Linux.gitignore: KDE directory preferences
	".directory",

	// From Global/KDevelop4.gitignore
	".kdev4",

	// From Global/JetBrains.gitignore
	".idea",

	// From Global/Eclipse.gitignore
	".settings",

	// From Global/Dropbox.gitignore
	".dropbox",
	".dropbox.attr",
	".dropbox.cache",

	// From Global/Cloud9.gitignore
	".c9revisions",
	".c9",

	// From Global/Bazaar.gitignore
	".bzr",
	".bzrignore",
}
