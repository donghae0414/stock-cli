package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Source int

const (
	SourceNone Source = iota
	SourceFile
)

const MissingCredentialsMessage = "missing Kiwoom credentials: run 'stock config set'"

func (s Source) String() string {
	switch s {
	case SourceFile:
		return "file"
	default:
		return "none"
	}
}

type Credentials struct {
	AppKey          string
	SecretKey       string
	AppKeySource    Source
	SecretKeySource Source
}

func (c Credentials) IsEmpty() bool {
	return c.AppKey == "" && c.SecretKey == ""
}

type configFile struct {
	Kiwoom struct {
		AppKey    string `toml:"appkey"`
		SecretKey string `toml:"secretkey"`
	} `toml:"kiwoom"`
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".stock", "config"), nil
}

func Load() (Credentials, error) {
	return LoadFrom(os.UserHomeDir)
}

func LoadFrom(homeDir func() (string, error)) (Credentials, error) {
	home, err := homeDir()
	if err != nil {
		return Credentials{}, fmt.Errorf("could not determine home directory: %w", err)
	}

	configPath := filepath.Join(home, ".stock", "config")
	file, fileErr := readFile(configPath)
	if fileErr != nil {
		return Credentials{}, fileErr
	}

	var creds Credentials

	if file.Kiwoom.AppKey != "" {
		creds.AppKey = file.Kiwoom.AppKey
		creds.AppKeySource = SourceFile
	}

	if file.Kiwoom.SecretKey != "" {
		creds.SecretKey = file.Kiwoom.SecretKey
		creds.SecretKeySource = SourceFile
	}

	return creds, nil
}

func LoadFile() (Credentials, error) {
	path, err := Path()
	if err != nil {
		return Credentials{}, err
	}
	file, err := readFile(path)
	if err != nil {
		return Credentials{}, err
	}

	var creds Credentials
	if file.Kiwoom.AppKey != "" {
		creds.AppKey = file.Kiwoom.AppKey
		creds.AppKeySource = SourceFile
	}
	if file.Kiwoom.SecretKey != "" {
		creds.SecretKey = file.Kiwoom.SecretKey
		creds.SecretKeySource = SourceFile
	}
	return creds, nil
}

func Save(c Credentials) error {
	path, err := Path()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("could not create directory %s: %w", dir, err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0700); err != nil {
			return fmt.Errorf("could not set permissions on %s: %w", dir, err)
		}
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("could not write to %s: %w", path, err)
	}
	defer f.Close()

	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("could not set permissions on %s: %w", path, err)
	}

	header := "# Stock CLI Kiwoom credentials.\n# Managed by `stock config set`. Do not edit manually unless you know what you are doing.\n\n"
	if _, err := f.WriteString(header); err != nil {
		return err
	}

	cfg := configFile{}
	cfg.Kiwoom.AppKey = c.AppKey
	cfg.Kiwoom.SecretKey = c.SecretKey

	return toml.NewEncoder(f).Encode(cfg)
}

func readFile(path string) (configFile, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return configFile{}, nil
	}
	if err != nil {
		return configFile{}, fmt.Errorf("could not open %s: %w", path, err)
	}
	defer f.Close()

	if runtime.GOOS != "windows" {
		if info, statErr := f.Stat(); statErr == nil {
			if perm := info.Mode().Perm(); perm&0077 != 0 {
				fmt.Fprintf(os.Stderr, "warning: %s has loose permissions (found: 0%o, expected: 0600)\n", path, perm)
			}
		}
	}

	var cfg configFile
	if _, err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		return configFile{}, fmt.Errorf("failed to parse %s: %w (try running 'stock config set' to re-create it)", path, err)
	}
	return cfg, nil
}
