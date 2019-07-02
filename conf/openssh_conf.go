package conf

import (
	"os"
	"regexp"

	"github.com/blacknon/lssh/common"
	"github.com/kevinburke/ssh_config"
)

func openOpenSshConfig(path string) (cfg *ssh_config.Config, err error) {
	// Read Openssh Config
	sshConfigFile := common.GetFullPath(path)
	f, err := os.Open(sshConfigFile)
	if err != nil {
		return
	}

	cfg, err = ssh_config.Decode(f)
	return
}

func getOpenSshConfig(path string) (config map[string]ServerConfig, err error) {
	config = map[string]ServerConfig{}

	// open openssh config
	cfg, err := openOpenSshConfig(path)
	if err != nil {
		return
	}

	// Get Node names
	hostList := []string{}
	for _, h := range cfg.Hosts {
		// not supported wildcard host
		re := regexp.MustCompile("\\*")
		for _, pattern := range h.Patterns {
			if !re.MatchString(pattern.String()) {
				hostList = append(hostList, pattern.String())
			}
		}
	}

	// append ServerConfig
	for _, host := range hostList {
		serverConfig := ServerConfig{
			Addr:         ssh_config.Get(host, "HostName"),
			Port:         ssh_config.Get(host, "Port"),
			User:         ssh_config.Get(host, "User"),
			ProxyCommand: ssh_config.Get(host, "ProxyCommand"),
			PreCmd:       ssh_config.Get(host, "LocalCommand"),
			Note:         "from :" + path,
		}

		// @TODO:
		//     Certificateは複数指定可能なようだが、現状は複数のファイルを受け付けるように作っていない。
		//     とりあえずそんなに使うことも無いだろうから、一旦単独ファイルとして渡させる。
		key := ssh_config.Get(host, "IdentityFile")
		cert := ssh_config.Get(host, "Certificate")
		if cert != "" {
			serverConfig.Cert = cert
			serverConfig.CertKey = key
		} else {
			serverConfig.Key = key
		}

		pkcs11Provider := ssh_config.Get(host, "PKCS11Provider")
		if pkcs11Provider != "" {
			serverConfig.PKCS11Use = true
			serverConfig.PKCS11Provider = pkcs11Provider
		}

		serverName := path + ":" + host
		config[serverName] = serverConfig
	}

	return config, err
}