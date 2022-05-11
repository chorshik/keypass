package manifest

var wrapperTemplate = `#!/bin/bash

if [ -f ~/.gpg-agent-info ] && [ -n "$(pgrep gpg-agent)" ]; then
source ~/.gpg-agent-info
export GPG_AGENT_INFO
else
eval $(gpg-agent --daemon)
fi

export PATH="$PATH:/usr/local/bin" # required on MacOS/brew
export GPG_TTY="$(tty)"
%s jsonapi listen
exit $?`

// DefaultBrowser ...
var DefaultBrowser = "firefox"

// DefaultWrapperPath ...
var DefaultWrapperPath = "/usr/local/bin"

// ValidBrowsers ...
var ValidBrowsers = []string{"chrome", "chromium", "firefox"}

var name = "com.github.ebladroher.native"
var wrapperName = "keypass_wrapper.sh"
var description = "Keypass wrapper оболочка для поиска и возврата паролей"
var connectionType = "stdio"
var chromeOrigins = []string{
	"chrome-extension://",
}
var firefoxOrigins = []string{
	"{27d32d38-b9ec-4fc4-8892-ae4b0dcb57c4}",
}

type manifestBase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Type        string `json:"type"`
}

type chromeManifest struct {
	manifestBase
	AllowedOrigins []string `json:"allowed_origins"`
}

type firefoxManifest struct {
	manifestBase
	AllowedExtensions []string `json:"allowed_extensions"`
}
