#!/bin/bash

# Install environment for development on the WeDeploy CLI Tool

set -euo pipefail
IFS=$'\n\t'

function checkCONT() {
  if [[ $CONT != "y" && $CONT != "yes" ]]; then
    exit 1
  fi
}

function passGit() {
  (which git >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    return
  fi

  (which xcode-select >> /dev/null) && ecbrew=$? || ecbrew=$?
  if [ $ecxcode -ne 0 ] ; then
    >&2 echo "Git wasn't found. Install it with your package manager or download it from"
    >&2 echo "https://git-scm.com"
    exit 1
  fi

  read -p "Git wasn't found. Install Command Lines Tools using Xcode? [no]: " CONT < /dev/tty
  checkCONT
  xcode-select --install
}

function passGo() {
  (which go >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    return
  fi

  if [ $ecbrew -ne 0 ] ; then
    >&2 echo "Go wasn't found. Install it with your package manager or download it from"
    >&2 echo "https://golang.org"
    exit 1
  fi

  read -p "Go wasn't found. Install Go with brew? [no]: " CONT < /dev/tty
  checkCONT
  brew install go
}

function passGoVisualCodeExtension() {
  (which code >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    installGoVisualCodeExtension
    return
  fi

  tipGoVisualCode
}

function installGoVisualCodeExtension() {
  echo "Installing Go extension for Visual Studio Code."
  # don't ask for doing 'Go: Install/Update Tools' on VS because the user is probably going to be prompted soon
  code --install-extension lukehoban.Go

  if [ $ecbrew -eq 0 ] ; then
    maybeInstallDelveOnMac
  fi
}

function maybeInstallDelveOnMac() {
  (which dlv >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    echo "Skipping installing debugger (Delver already installed)."
    return
  fi

  read -p "Installing the debugger adds a self-signed certificate on your keychain. Continue? [no]: " CONT < /dev/tty
  checkCONT

  echo "Installing debugger for macOS"
  curl https://raw.githubusercontent.com/derekparker/delve/master/scripts/gencert.sh | bash
  brew install go-delve/delve/delve
}

function tipGoVisualCode() {
  echo "[INFO] Tip: You might want to consider using Visual Studio Code for working with Go code."
  echo "https://code.visualstudio.com/"
  echo "If you already use it, open the Command Palette and select \"Shell Command: Install 'code' command in PATH\""
  echo "Then use it to install the Go extension or install it using the CLI:"
  echo "\"code --install-extension lukehoban.Go\""

  if [ $UNAME == "darwin" ] ; then
    echo "For macOS, you must install the Go debugger for Visual Studio Code manually."
    echo "Please see https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code"
    echo "and https://github.com/go-delve/homebrew-delve/issues/19"
    echo "Then run the following commands:"
    echo "curl https://raw.githubusercontent.com/derekparker/delve/master/scripts/gencert.sh | bash"
    echo "brew install go-delve/delve/delve"
    # on other systems it is already installed automatically
    # https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code
  fi

  read -p "Continue? [no]: " CONT < /dev/tty
  checkCONT
}

function setupGopath() {
  GOPATH=${GOPATH:-""}
  
  if [ ! -z $GOPATH ] ; then
    echo "Skipping setting \$GOPATH. \$GOPATH is already set to $GOPATH"
    return
  fi

  echo "You must set the \$GOPATH environment variable now."
  echo "More information on https://golang.org/doc/code.html#GOPATH"
  echo
  echo "GOPATH is the location where your Go ecosystem/files should live (including this repository)."
  echo "After you set GOPATH the following directory \$GOPATH/bin will also be added on your \$PATH"
  read -p "Set GOPATH [default: ~/go]: " gp < /dev/tty;
  GP_TMP_SET=true
  GP_TEMP=${GP_TEMP:-$HOME/go}
  export GOPATH=$GP_TEMP
  
  if [ -f $HOME/.bash_profile ] ; then
    echo "export GOPATH=$GP_TEMP" >> $HOME/.zshrc
  fi

  if [ -f $HOME/.bashrc ] ; then
    echo "export GOPATH=$GP_TEMP" >> $HOME/.bashrc
  fi

  if [ -f $HOME/.zshrc ] ; then
    echo "export GOPATH=$GP_TEMP" >> $HOME/.zshrc
  fi

  if [ -f $HOME/.config/fish/config.fish ] ; then
    echo "set -x GOPATH $GP_TEMP" >> $HOME/.config/fish/config.fish
  fi

  mkdir -p $GOPATH/bin
  mkdir -p $GOPATH/src/github.com/wedeploy

  case ":$PATH:" in
    *:$GP_TEMP/bin:*) return;;
    *)
      export PATH=$GP_TEMP/bin:\$PATH

      if [ -f $HOME/.bash_profile ] ; then
        echo "export PATH=$GP_TEMP/bin:\$PATH" >> $HOME/.bash_profile
      fi

      if [ -f $HOME/.bashrc ] ; then
        echo "export PATH=$GP_TEMP/bin:\$PATH" >> $HOME/.bashrc
      fi

      if [ -f $HOME/.zshrc ] ; then
        echo "export PATH=$GP_TEMP/bin:\$PATH" >> $HOME/.zshrc
      fi

      if [ -f $HOME/.config/fish/config.fish ] ; then
        echo "set -x PATH $GP_TEMP/bin:\$PATH" >> $HOME/.config/fish/config.fish
      fi
    ;;
  esac
}

function setupIAlias() {
  (which i >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    echo "Skipping setting alias \"i\". It is already in use"
    which i
    return
  fi

  GP_TMP_SET=true

  if [ -f $HOME/.bash_profile ] ; then
    echo "alias i=\"\$GOPATH/src/github.com/wedeploy/cli/scripts/build-run.sh \$@\"" >> $HOME/.bash_profile
  fi

  if [ -f $HOME/.bashrc ] ; then
    echo "alias i=\"\$GOPATH/src/github.com/wedeploy/cli/scripts/build-run.sh \$@\"" >> $HOME/.bashrc
  fi

  if [ -f $HOME/.zshrc ] ; then
    echo "alias i=\"\$GOPATH/src/github.com/wedeploy/cli/scripts/build-run.sh \$@\"" >> $HOME/.zshrc
  fi

  if [ -f $HOME/.config/fish/config.fish ] ; then
    echo "alias i=\"\$GOPATH/src/github.com/wedeploy/cli/scripts/build-run.sh \$argv[2..-1]\"" >> $HOME/.config/fish/config.fish
  fi
}

function passGoDevDependencies() {
  echo "Installing developer dependencies:"
  
  echo "vendorlicenses (github.com/henvic/vendorlicenses)"
  go get -u github.com/henvic/vendorlicenses
  
  echo "errcheck (github.com/kisielk/errcheck)"
  go get -u github.com/kisielk/errcheck
  
  echo "golint (github.com/golang/lint/golint)"
  go get -u github.com/golang/lint/golint
  
  echo "megacheck (https://github.com/dominikh/go-tools/tree/master/cmd/megacheck)"
  go get -u honnef.co/go/tools/cmd/megacheck
}

function maybeMoveToGopathDir() {
  BN=$(basename $PWD)
  WEDEPLOY_GO_DIR=$GOPATH/src/github.com/wedeploy

  if [ $PWD = $WEDEPLOY_GO_DIR/cli ] ; then
    return
  fi

  echo "You need to keep Go code together inside your \$GOPATH respecting Go package paths schema."
  echo "Therefore, the CLI package is going to be moved from $PWD to $WEDEPLOY_GO_DIR."
  read -p "Continue? [no]: " CONT < /dev/tty
  checkCONT
  
  if [ -f $WEDEPLOY_GO_DIR/cli ] ; then
    >&2 echo "A directory already exists for this package on your \$GOPATH: $WEDEPLOY_GO_DIR/cli. Aborting."
    exit 1
  fi

  mv ../$BN $WEDEPLOY_GO_DIR/cli
  echo "Project moved to $WEDEPLOY_GO_DIR/cli"
}

function infoRenewShell() {
  if [ $GP_TMP_SET == true ] ; then
    echo "The following environment variables \$GOPATH and \$PATH and the alias \"i\" were created or modified."
    echo "You must now either close and reopen all shells to start using them or call these commands:"
    echo "export GOPATH=$GOPATH"
    echo "export PATH=$GOPATH/bin:\$PATH"
    echo "alias i=\"\$GOPATH/src/github.com/wedeploy/cli/scripts/build-run.sh \$@\""
  fi

  echo
  echo "Once you are ready try to compile the software by running \"go build\" and then executing \"./cli\""
  echo "To compile and immediately run the CLI program you can run \"i\" instead of \"we\" in any directory on your shell."
}

GP_TMP=$HOME/go
GP_TMP_SET=false
UNAME=$(uname | tr '[:upper:]' '[:lower:]')
cd $(dirname $0)/..
(which brew >> /dev/null) && ecbrew=$? || ecbrew=$?

passGit
passGo
echo "If you use the Vim editor, see https://github.com/fatih/vim-go"
passGoVisualCodeExtension
setupGopath
setupIAlias
passGoDevDependencies
maybeMoveToGopathDir
infoRenewShell
