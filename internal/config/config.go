package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/pbkdf2"
)

const (
	CONFIG_EXTENSION  = ".toml"
	SECURE_EXTENSION  = ".secure"
	SALT_LENGTH       = 32
	PBKDF2_ITERATIONS = 310_000
)

type Config struct {
	Server      string `toml:"server"`
	Port        string `toml:"port"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
	DisplayName string `toml:"display_name"`
}

// ListConfigs lists all configurations in the basePath.
// It returns a map with [key: {configPath - extensions} value: encrypted?].
func ListConfigs(basePath string) (map[string]bool, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	configs := map[string]bool{}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), CONFIG_EXTENSION+SECURE_EXTENSION) {
			configs[strings.TrimSuffix(
				filepath.Join(basePath, entry.Name()), CONFIG_EXTENSION+SECURE_EXTENSION,
			)] = true
		} else if strings.HasSuffix(entry.Name(), CONFIG_EXTENSION) {
			configs[strings.TrimSuffix(
				filepath.Join(basePath, entry.Name()), CONFIG_EXTENSION,
			)] = false
		}
	}
	return configs, nil
}

// LoadConfig loads and deserializes the configuration from disk.
// Fileextensions are added automatically.
// If a decryptionKey is set, the configuration is decrypted with aes256.
func LoadConfig(path, decryptionKey string) (*Config, error) {
	path += CONFIG_EXTENSION
	if decryptionKey != "" {
		path += SECURE_EXTENSION
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rawConfig, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if decryptionKey != "" {
		if len(rawConfig) < SALT_LENGTH {
			return nil, fmt.Errorf("invalid ciphertext")
		}
		salt, ciphertext := rawConfig[:SALT_LENGTH], rawConfig[SALT_LENGTH:]

		block, err := aes.NewCipher(pbkdf2.Key(
			[]byte(decryptionKey), salt, PBKDF2_ITERATIONS, SALT_LENGTH, sha256.New,
		))
		if err != nil {
			return nil, err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		if len(ciphertext) < gcm.NonceSize() {
			return nil, fmt.Errorf("invalid ciphertext")
		}
		rawConfig, err = gcm.Open(nil, ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():], nil)
		if err != nil {
			return nil, err
		}
	}

	config := &Config{}
	err = toml.Unmarshal(rawConfig, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// WriteConfig serializes the configuration and writes it to disk.
// Fileextensions are added automatically.
// If an encryptionKey is set, the configuration is encrypted with aes256.
func WriteConfig(config *Config, path, encryptionKey string) error {
	path += CONFIG_EXTENSION
	rawConfig, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	if encryptionKey != "" {
		path += SECURE_EXTENSION

		salt := make([]byte, SALT_LENGTH)
		_, err := rand.Read(salt)
		if err != nil {
			return fmt.Errorf("failed to generate salt: %w", err)
		}

		block, err := aes.NewCipher(pbkdf2.Key(
			[]byte(encryptionKey), salt, PBKDF2_ITERATIONS, SALT_LENGTH, sha256.New,
		))
		if err != nil {
			return err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return err
		}

		nonce := make([]byte, gcm.NonceSize())
		_, err = rand.Read(nonce)
		if err != nil {
			return fmt.Errorf("failed to generate nonce: %w", err)
		}

		rawConfig = gcm.Seal(nil, nonce, rawConfig, nil)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(rawConfig)
	if err != nil {
		return err
	}

	return nil
}

// RemoveConfig deletes a configuration from disk.
// Fileextensions are added automatically.
func RemoveConfig(path string, encrypted bool) error {
	path += CONFIG_EXTENSION
	if encrypted {
		path += SECURE_EXTENSION
	}

	return os.Remove(path)
}
