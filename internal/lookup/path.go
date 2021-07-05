/*
# Copyright (c) 2021, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
*/

package lookup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	envPath = "PATH"
)

var defaultPaths = []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin"}

type path struct {
	file
}

func NewPathLocator(root string) Locator {
	return NewPathLocatorWithLogger(log.StandardLogger(), root)
}

func NewPathLocatorWithLogger(logger *log.Logger, root string) Locator {
	pathEnv := os.Getenv(envPath)
	paths := filepath.SplitList(pathEnv)

	if root != "" {
		paths = append(paths, defaultPaths...)
	}

	var prefixes []string
	for _, dir := range paths {
		prefixes = append(prefixes, filepath.Join(root, dir))
	}
	l := path{
		file: file{
			logger:   logger,
			prefixes: prefixes,
			filter:   assertExecutable,
		},
	}
	return &l
}

var _ Locator = (*path)(nil)

func (p path) Locate(filename string) ([]string, error) {
	// For absolute paths we ensure that it is executable
	if strings.Contains(filename, "/") {
		err := assertExecutable(filename)
		if err != nil {
			return nil, fmt.Errorf("absolute path %v is not an executable file: %v", filename, err)
		}
		return []string{filename}, nil
	}

	return p.file.Locate(filename)
}

func assertExecutable(filename string) error {
	err := assertFile(filename)
	if err != nil {
		return err
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("specified file '%v' is not executable", filename)
	}

	return nil
}
