
package main

import (
    "os"
    "fmt"
    "encoding/json"
    "log"
)

type Config struct {
	RootDir    string
	PublicKey  string
	PrivateKey string
	Name       string
	Email      string
}

func NewConfig(path string) (*Config, error) {
    log.Printf("Opening config file: %s", path)
    fp, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("error opening config file %v", err)
    }
    defer fp.Close()

    var buf []byte = make([]byte, 4096)
    n, err := fp.Read(buf)
    buf = buf[:n]
    if err != nil {
        return nil, fmt.Errorf("error reading config file: %v", err)
    }

    var config Config
    err = json.Unmarshal(buf, &config)
    if err != nil {
        return nil, fmt.Errorf("invalid json: %v", err)
    }

    return &config, nil
}

