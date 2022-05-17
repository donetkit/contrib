package config

import (
	"errors"
	"github.com/donetkit/contrib/utils/files"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

type YMLConfig struct {
}

func (c *YMLConfig) LoadFullPath(filePath string, data interface{}) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if err = yaml.NewDecoder(file).Decode(data); err != nil {
		return err
	}
	return nil
}

func (c *YMLConfig) SaveFullPath(filePath string, data interface{}) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if err := yaml.NewEncoder(file).Encode(data); err != nil {
		return err
	}
	return nil
}

func (c *YMLConfig) Load(filePath, fileName string, data interface{}, save ...bool) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	if fileName == "" {
		return errors.New("file name is empty")
	}
	file, err := os.Open(path.Join(files.CurrentPath, filePath, fileName))
	if err != nil {
		if len(save) > 0 && save[0] {
			c.Save(filePath, fileName, data)
		}
		return err
	}
	defer func() { _ = file.Close() }()
	if err = yaml.NewDecoder(file).Decode(data); err != nil {
		return err
	}
	return nil
}

func (c *YMLConfig) Save(filePath, fileName string, data interface{}) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	if fileName == "" {
		return errors.New("file name is empty")
	}
	files.IsNotExistMkDir(path.Join(files.CurrentPath, filePath))
	file, err := os.Create(path.Join(files.CurrentPath, filePath, fileName))
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if err := yaml.NewEncoder(file).Encode(data); err != nil {
		return err
	}
	return nil
}
