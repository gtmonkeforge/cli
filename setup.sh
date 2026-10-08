#!/bin/bash
# curl -sL "https://monkeforge.org/cli/setup.sh" | bash

OutputDir="$HOME/.local/share/mforge"

if [ ! -d "$OutputDir" ]; then
    mkdir -p "$OutputDir"
fi

curl -L "https://github.com/sirkingbinx/monkeforge-cli/releases/latest/download/mforge-linux-x64" -o "$OutputDir/mforge"

chmod +x "$OutputDir/mforge"

if [[ ":$PATH:" != *":$OutputDir:"* ]]; then
    echo "export PATH=\"\$PATH:$OutputDir\"" >> "$HOME/.bashrc"
fi

export PATH="$PATH:$OutputDir"

# We're done
echo "'mforge' command is installed"
echo "Terminal must be restarted for changes to take effect