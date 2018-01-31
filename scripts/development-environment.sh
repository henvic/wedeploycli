#!/bin/bash

# Install environment for development on the WeDeploy CLI Tool

set -euo pipefail
IFS=$'\n\t'

function checkNoCONT() {
  if [[ $WE_DEVENV_CONT == "n" || $WE_DEVENV_CONT == "no" ]]; then
    return
  fi

  exit 1
}

function checkYesCONT() {
  if [[ $WE_DEVENV_CONT == "y" || $WE_DEVENV_CONT == "yes" ]]; then
    return
  fi

  exit 1
}

function welcome() {
  echo "WeDeploy Command-Line-Interface Development Environment installer."
  echo ""
  echo "This program tries to install and verify all necessary dependencies for doing CLI development."
  echo "It should work on Linux and macOS and either guide the user through the instalation process."
  echo ""
}

function passGit() {
  (which git >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    return
  fi

  (which xcode-select >> /dev/null) && WE_DEV_ENV_ECBREW=$? || WE_DEV_ENV_ECBREW=$?
  if [ $ecxcode -ne 0 ] ; then
    >&2 echo "Git wasn't found. Install it with your package manager or download it from"
    >&2 echo "https://git-scm.com"
    exit 1
  fi

  read -p "Git wasn't found. Install Command Lines Tools using Xcode? [yes]: " WE_DEVENV_CONT < /dev/tty
  checkYesCONT
  xcode-select --install
}

function passGo() {
  (which go >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    return
  fi

  if [ $WE_DEV_ENV_ECBREW -ne 0 ] ; then
    >&2 echo "Go wasn't found. Install it with your package manager or download it from"
    >&2 echo "https://golang.org"
    exit 1
  fi

  read -p "Go wasn't found. Install Go with brew? [yes]: " WE_DEVENV_CONT < /dev/tty
  checkYesCONT
  brew install go
}

function passVimExtension() {
  echo "If you use the Vim editor, see https://github.com/fatih/vim-go"
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
  (code --list-extensions | grep ^lukehoban.Go$ >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    echo "Skipping installing Visual Studio Code extension for Go (already installed)."
    return
  fi

  echo "Installing Go extension for Visual Studio Code."
  code --install-extension lukehoban.Go
  # don't ask for doing 'Go: Install/Update Tools' on VS because the user is probably going to be prompted soon

  if [ $ec -eq 0 ] ; then
    maybeInstallDelveOnMac
  fi
}

function maybeInstallDelveOnMac() {
  (which dlv >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    echo "Skipping installing debugger (Delver already installed)."
    return
  fi

  read -p "Installing the debugger adds a self-signed certificate on your keychain. Continue? [yes]: " WE_DEVENV_CONT < /dev/tty
  checkYesCONT

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

  if [ $WE_DEVENV_UNAME == "darwin" ] ; then
    echo "For macOS, you must install the Go debugger for Visual Studio Code manually."
    echo "Please see https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code"
    echo "and https://github.com/go-delve/homebrew-delve/issues/19"
    echo "Then run the following commands:"
    echo "curl https://raw.githubusercontent.com/derekparker/delve/master/scripts/gencert.sh | bash"
    echo "brew install go-delve/delve/delve"
    # on other systems it is already installed automatically
    # https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code
  fi

  read -p "Continue? [yes]: " WE_DEVENV_CONT < /dev/tty
  checkYesCONT
}

function setupGopath() {
  export GOPATH=${GOPATH:-""}
  
  if [ ! -z $GOPATH ] ; then
    echo "Skipping setting \$GOPATH. \$GOPATH is already set to $GOPATH"
    return
  fi

  echo "You must set the \$GOPATH environment variable now."
  echo "More information on https://golang.org/doc/code.html#GOPATH"
  echo
  echo "GOPATH is the location where your Go ecosystem/files should live (including this repository)."
  echo "After you set GOPATH the following directory \$GOPATH/bin will also be added on your \$PATH"
  read -p "Set \$GOPATH [default: ~/go]: " gp < /dev/tty;
  WE_DEVENV_NEW_GOPATH_SET=true
  export GOPATH=${GOPATH:-$HOME/go}
  
  echo "export GOPATH=$GOPATH" >> $HOME/.bashrc

  if [ -f $HOME/.zshrc ] ; then
    echo "export GOPATH=$GOPATH" >> $HOME/.zshrc
  fi

  if [ -f $HOME/.config/fish/config.fish ] ; then
    echo "set -x GOPATH $GOPATH" >> $HOME/.config/fish/config.fish
  fi

  mkdir -p $GOPATH/bin
  mkdir -p $GOPATH/src/github.com/wedeploy

  case ":$PATH:" in
    *:$GOPATH/bin:*) return;;
    *)
      export PATH=$GOPATH/bin:$PATH

      echo "export PATH=$GOPATH/bin:\$PATH" >> $HOME/.bashrc

      if [ -f $HOME/.zshrc ] ; then
        echo "export PATH=$GOPATH/bin:\$PATH" >> $HOME/.zshrc
      fi

      if [ -f $HOME/.config/fish/config.fish ] ; then
        echo "set -x PATH $GOPATH/bin:\$PATH" >> $HOME/.config/fish/config.fish
      fi
    ;;
  esac
}

function setupI() {
  if [ ! -f $GOPATH/bin/i ] ; then
    ln -s $GOPATH/src/github.com/wedeploy/cli/scripts/build-run.sh $GOPATH/bin/i
  fi
}

function passGoDevDependencies() {
  echo "Installing developer tools."
  
  echo "vendorlicenses https://github.com/henvic/vendorlicenses"
  go get -u github.com/henvic/vendorlicenses
  
  echo "errcheck https://github.com/kisielk/errcheck"
  go get -u github.com/kisielk/errcheck
  
  echo "golint https://github.com/golang/lint/golint"
  go get -u github.com/golang/lint/golint
  
  echo "megacheck https://github.com/dominikh/go-tools/tree/master/cmd/megacheck"
  go get -u honnef.co/go/tools/cmd/megacheck
}

function maybeMoveToGopathDir() {
  WE_DEVENV_BN=$(basename $PWD)
  WE_DEVENV_GO_DIR=$GOPATH/src/github.com/wedeploy

  if [ $PWD = $WE_DEVENV_GO_DIR/cli ] ; then
    return
  fi

  echo "You need to keep Go code together inside your \$GOPATH respecting Go package paths schema."
  echo "Therefore, the CLI package is going to be moved from $PWD to $WE_DEVENV_GO_DIR."
  read -p "Continue? [yes]: " WE_DEVENV_CONT < /dev/tty
  checkYesCONT
  
  if [ -f $WE_DEVENV_GO_DIR/cli ] ; then
    >&2 echo "A directory already exists for this package on your \$GOPATH: $WE_DEVENV_GO_DIR/cli. Aborting."
    exit 1
  fi

  mv ../$WE_DEVENV_BN $WE_DEVENV_GO_DIR/cli
  echo "Project moved to $WE_DEVENV_GO_DIR/cli"
}

function infoRenewShell() {
  echo "Compile and immediately run the CLI program by using \"i\" instead of \"we\" from inside any directory."
  echo "For example: instead of using \"we list\" to list your services, use \"i list\" like this:"
  echo
  echo "$ i list --help"
  bash -c "i deploy --help"

  if [[ $WE_DEVENV_NEW_GOPATH_SET == true ]] ; then
    echo
    echo "The following environment variables \$GOPATH and \$PATH were created:"
    echo "export GOPATH=$GOPATH"
    echo "export PATH=$GOPATH/bin:\$PATH"
    echo
    echo "You might need to open a new shell after this to start using the \"i\" command."
  fi
}

WE_DEVENV_NEW_GOPATH_SET=false
WE_DEVENV_UNAME=$(uname | tr '[:upper:]' '[:lower:]')
cd $(dirname $0)/..
(which brew >> /dev/null) && WE_DEV_ENV_ECBREW=$? || WE_DEV_ENV_ECBREW=$?

welcome
passGit
passGo
passVimExtension
passGoVisualCodeExtension
setupGopath
setupI
passGoDevDependencies
maybeMoveToGopathDir
infoRenewShell
