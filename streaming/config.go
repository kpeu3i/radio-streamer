package streaming

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Streams       []string `yaml:"streams"`
	CurrentStream int      `yaml:"current_stream"`
	CurrentVolume string   `yaml:"current_volume"`
}

type ConfigFileStorage struct {
	filename string
}

func NewConfigStorage(filename string) *ConfigFileStorage {
	return &ConfigFileStorage{filename: filename}
}

func (s *ConfigFileStorage) Load() (Config, error) {
	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return Config{}, err
	}

	defer func() {
		_ = file.Close()
	}()

	config := Config{}

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil && err != io.EOF {
		return Config{}, err
	}

	return config, nil
}

func (s *ConfigFileStorage) Store(config Config) error {
	file, err := os.OpenFile(s.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	decoder := yaml.NewEncoder(file)
	err = decoder.Encode(config)
	if err != nil {
		return err
	}

	return nil
}
