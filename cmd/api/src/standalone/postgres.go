// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package standalone owns the local PostgreSQL process used by the standalone
// BHCE distribution.
package standalone

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	databaseName             = "bloodhound"
	databaseUser             = "bloodhound"
	standaloneMaxConnections = 100
)

type Postgres struct {
	binDirectory  string
	dataDirectory string
	port          uint16
	password      string
}

func NewPostgres(binDirectory string, dataDirectory string, port uint16, password string) Postgres {
	return Postgres{
		binDirectory:  binDirectory,
		dataDirectory: dataDirectory,
		port:          port,
		password:      password,
	}
}

func AllocatePort() (uint16, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("allocating local PostgreSQL port: %w", err)
	}
	defer listener.Close()

	portString := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("parsing allocated local PostgreSQL port: %w", err)
	}

	return uint16(port), nil
}

func (s Postgres) Start() error {
	if err := s.validateBinaries(); err != nil {
		return err
	}

	initialized, err := s.initialize()
	if err != nil {
		return err
	}

	if err := s.run("pg_ctl", "start", "-w", "-D", s.dataDirectory, "-o", strings.Join([]string{
		"-p", strconv.Itoa(int(s.port)),
		"-h", "127.0.0.1",
		"-c", "listen_addresses=127.0.0.1",
		"-c", "max_connections=" + strconv.Itoa(standaloneMaxConnections),
	}, " ")); err != nil {
		return err
	}

	if initialized {
		if err := s.runWithPassword("createdb", "--host=127.0.0.1", "--port="+strconv.Itoa(int(s.port)), "--username="+databaseUser, databaseName); err != nil {
			_ = s.Stop()
			return fmt.Errorf("creating standalone database: %w", err)
		}
	}

	return nil
}

func (s Postgres) Stop() error {
	if _, err := os.Stat(filepath.Join(s.dataDirectory, "postmaster.pid")); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("checking PostgreSQL state: %w", err)
	}

	if err := s.run("pg_ctl", "stop", "-w", "-m", "fast", "-D", s.dataDirectory); err != nil {
		return fmt.Errorf("stopping embedded PostgreSQL: %w", err)
	}

	return nil
}

func (s Postgres) initialize() (bool, error) {
	if _, err := os.Stat(filepath.Join(s.dataDirectory, "PG_VERSION")); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("checking PostgreSQL data directory: %w", err)
	}

	if err := os.MkdirAll(s.dataDirectory, 0o700); err != nil {
		return false, fmt.Errorf("creating PostgreSQL data directory: %w", err)
	}

	passwordFile, err := os.CreateTemp(filepath.Dir(s.dataDirectory), "initdb-password-*")
	if err != nil {
		return false, fmt.Errorf("creating PostgreSQL password file: %w", err)
	}
	passwordFilePath := passwordFile.Name()
	defer os.Remove(passwordFilePath)

	if _, err := passwordFile.WriteString(s.password); err != nil {
		passwordFile.Close()
		return false, fmt.Errorf("writing PostgreSQL password file: %w", err)
	}
	if err := passwordFile.Close(); err != nil {
		return false, fmt.Errorf("closing PostgreSQL password file: %w", err)
	}

	if err := s.run("initdb", "-D", s.dataDirectory, "--username="+databaseUser, "--pwfile="+passwordFilePath, "--auth=scram-sha-256", "--encoding=UTF8"); err != nil {
		return false, err
	}

	return true, nil
}

func (s Postgres) validateBinaries() error {
	for _, binaryName := range []string{"createdb", "initdb", "pg_ctl"} {
		binaryPath := filepath.Join(s.binDirectory, binaryName)
		if _, err := os.Stat(binaryPath); err != nil {
			return fmt.Errorf("standalone PostgreSQL runtime is incomplete: expected %s: %w", binaryPath, err)
		}
	}

	return nil
}

func (s Postgres) run(binaryName string, arguments ...string) error {
	command := exec.Command(filepath.Join(s.binDirectory, binaryName), arguments...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("running %s: %w", binaryName, err)
	}

	return nil
}

func (s Postgres) runWithPassword(binaryName string, arguments ...string) error {
	command := exec.Command(filepath.Join(s.binDirectory, binaryName), arguments...)
	command.Env = append(os.Environ(), "PGPASSWORD="+s.password)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("running %s: %w", binaryName, err)
	}

	return nil
}
