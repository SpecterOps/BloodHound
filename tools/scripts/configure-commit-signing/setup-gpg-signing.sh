#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

# Create temporary dir
tmpdir=$(mktemp -d -t ci-$(date +%Y-m-%d-%H-%M-%S)-XXXXXXXXXX)

# Cleanup function
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Function to determine the shell and append export GPG_TTY=$(tty) to the appropriate rc file
configure_gpg_tty() {
  SHELL_NAME=$(basename "$SHELL")
  if [ "$SHELL_NAME" = "zsh" ]; then
    RC_FILE="$HOME/.zshrc"
  elif [ "$SHELL_NAME" = "bash" ]; then
    RC_FILE="$HOME/.bashrc"
  else
    echo "Unsupported shell: $SHELL_NAME"
    return
  fi

  # Append export GPG_TTY=$(tty) to the rc file if not already present
  if ! grep -qx 'export GPG_TTY=$(tty)' "$RC_FILE"; then
    echo 'export GPG_TTY=$(tty)' >>"$RC_FILE"
    echo "Added 'export GPG_TTY=\$(tty)' to $RC_FILE"
  else
    echo "'export GPG_TTY=\$(tty)' is already present in $RC_FILE"
  fi
}

# Call the function to configure GPG_TTY
configure_gpg_tty

brew_in() {
  if brew list --versions "$1" &>/dev/null; then
    brew upgrade "$1"
  else
    brew install "$1"
  fi
}

get_github_username() {
  username=$(gh auth status 2>&1 | awk -F 'account ' '/Logged in to github.com account / {print $2}' | awk '{print $1}')
  echo "$username"
}

# Check if Homebrew is installed, if not then install it
if ! command_exists brew; then
  echo "Homebrew is not installed. Installing Homebrew..."
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
else
  echo "Homebrew is already installed."
fi

# Update Homebrew
echo "Updating brew..."
brew update

# List of packages to install or upgrade
packages=(
  gpg2
  gnupg
  openssl
  pinentry-mac
  jq
  1password-cli
  gh
)

# Loop through the list and apply the function
for package in "${packages[@]}"; do
  echo "Install or updating $package"
  brew_in "$package"
done

# Ask for GitHub email, store it
read -p "Enter your GitHub email address associated with SpecterOps: " GITHUB_EMAIL

# Authenticate with 1Password using the email address
echo "Signing in to 1Password..."
eval "$(op signin)"

# Authenticate with GitHub
# echo "Authenticating with GitHub..."
gh auth login -h github.com -s admin:gpg_key

GITHUB_USERNAME=$(get_github_username)

GPG_DIR="${HOME}/.gnupg"
if [ ! -d "$GPG_DIR" ]; then
  echo "Creating $GPG_DIR with 700 permissions..."
  mkdir -p "$GPG_DIR"
  chmod 700 "$GPG_DIR"
else
  echo "$GPG_DIR already exists. Ensuring 700 permissions..."
  chmod 700 "$GPG_DIR"
fi

# Configure gpg-agent.conf
GPG_AGENT_CONF="${HOME}/.gnupg/gpg-agent.conf"
PINENTRY_LINE="pinentry-program $(brew --prefix)/bin/pinentry-mac"

if [ ! -f "$GPG_AGENT_CONF" ]; then
  echo "Creating $GPG_AGENT_CONF..."
  echo "$PINENTRY_LINE" >"$GPG_AGENT_CONF"
else
  echo "Updating $GPG_AGENT_CONF..."
  grep -qxF "$PINENTRY_LINE" "$GPG_AGENT_CONF" || echo "$PINENTRY_LINE" >>"$GPG_AGENT_CONF"
fi

# Configure gpg.conf
GPG_CONF="${HOME}/.gnupg/gpg.conf"
USE_AGENT_LINE="use-agent"
PINENTRY_CONF_LINE="pinentry-program $(brew --prefix)/bin/pinentry-mac"

if [ ! -f "$GPG_CONF" ]; then
  echo "Creating $GPG_CONF..."
  echo "$USE_AGENT_LINE" >"$GPG_CONF"
else
  echo "Updating $GPG_CONF..."
  grep -qxF "$USE_AGENT_LINE" "$GPG_CONF" || echo "$USE_AGENT_LINE" >>"$GPG_CONF"
fi

# Kill all gpg-agent processes
echo "Killing all gpg-agent processes..."
killall gpg-agent || true

echo "launching gpg agent"
gpgconf --launch gpg-agent

OP_ITEM_TITLE_GPG_PASSPHRASE="GPG Passphrase - $(uuidgen)"
OP_VAULT="Personal"

# Create a passphrase for GPG in 1Password
echo "Creating GPG passphrase in 1Password..."
op item create \
  --category "Password" \
  --title "$OP_ITEM_TITLE_GPG_PASSPHRASE" \
  --vault "$OP_VAULT" \
  --generate-password='letters,digits,64' \
  --tags gpg-key,github,passphrase

OP_GPG_PASSHPRASE_ID=$(op item get "$OP_ITEM_TITLE_GPG_PASSPHRASE" --format json | jq -r ".id")
GPG_PASSPHRASE=$(op item get $OP_GPG_PASSHPRASE_ID --reveal --fields password)

# Generate master key
gpg --batch --generate-key <<EOF
%echo Generating master key
Key-Type: eddsa
Key-Curve: ed25519
Key-Usage: cert
Name-Real: $GITHUB_USERNAME
Name-Email: $GITHUB_EMAIL
Name-Comment: "Generated from $(hostname)"
Passphrase: $GPG_PASSPHRASE
Expire-Date: 4y
%commit
%echo Master key generated
EOF

PRIMARY_KEY_FPR=$(gpg --list-keys --with-colons "$GITHUB_EMAIL" | awk -F: '/^pub:/ { pub=1 } /^fpr:/ && pub { fpr=$10; pub=0 } END { print fpr }')
echo "Generated main key with fingerprint: $PRIMARY_KEY_FPR"

echo "adding signing subkey"
gpg \
  --batch \
  --pinentry-mode=loopback \
  --passphrase="$GPG_PASSPHRASE" \
  --quick-add-key $PRIMARY_KEY_FPR ed25519 sign 1y

# Retrieve the long key IDs
PRIMARY_LONG_KEY_ID=$(gpg --with-colons --list-secret-keys "$GITHUB_EMAIL" | awk -F: '/^sec:/ { id=$5 } END { print id }')
SUBKEY_LONG_KEY_ID=$(gpg --with-colons --list-secret-keys "$GITHUB_EMAIL" | awk -F: '/^ssb:/ { id=$5 } END { print id }')

# Export the primary key
PRIMARY_SECRET_KEY=$(mktemp -p $tmpdir)
gpg --armor --export-secret-keys "$PRIMARY_LONG_KEY_ID" >"$PRIMARY_SECRET_KEY"
PRIMARY_PUBLIC_KEY=$(mktemp -p $tmpdir)
gpg --armor --export "$PRIMARY_LONG_KEY_ID" >"$PRIMARY_PUBLIC_KEY"
# Store the generated GPG keys in 1Password
echo "Storing GPG primary key in 1Password..."
cat $PRIMARY_SECRET_KEY |
  op document create - \
    --title "GPG Primary Secret Key" \
    --file-name "$GITHUB_USERNAME.priv.asc" \
    --tags gpg-key,private-key
cat $PRIMARY_PUBLIC_KEY |
  op document create - \
    --title "GPG Primary Public Key" \
    --file-name "$GITHUB_USERNAME.pub.asc" \
    --tags gpg-key,public-key
echo "Storing GPG subkey in 1Password..."
SUBKEY_PUBLIC_KEY=$(mktemp -p $tmpdir)
gpg --armor --export "$SUBKEY_LONG_KEY_ID" >"$SUBKEY_PUBLIC_KEY"
cat $SUBKEY_PUBLIC_KEY |
  op document create - \
    --title "GPG Subkey Public Key" \
    --file-name "$GITHUB_USERNAME.subkey.pub.asc" \
    --tags gpg-key,public-key

# Configure Git to use the subkey for signing
echo "Configuring Git to use GPG subkey for signing..."
git config --global user.signingkey "$SUBKEY_LONG_KEY_ID"
git config --global commit.gpgsign true
git config --global tag.gpgsign true
git config --global gpg.program "$(which gpg)"

# Add subkey's public key to GitHub
echo "Adding subkey's public key to GitHub..."
gh gpg-key add <(cat "$SUBKEY_PUBLIC_KEY") --title "GPG Subkey $SUBKEY_LONG_KEY_ID"

# Delete the primary key from local machine
echo "Deleting primary key from local machine..."
gpg --fingerprint --with-colons $GITHUB_EMAIL |
  grep "^fpr" |
  sed -n 's/^fpr:::::::::\([[:alnum:]]\+\):/\1/p' |
  xargs gpg --batch --delete-secret-keys | true

# Re-import the subkey
echo "Re-importing subkey..."
gpg --import "$SUBKEY_PUBLIC_KEY"

echo "Script completed successfully."
