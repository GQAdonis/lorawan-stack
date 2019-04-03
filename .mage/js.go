// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package ttnmage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func nodeBin(cmd string) string { return filepath.Join("node_modules", ".bin", cmd) }

// Js namespace.
type Js mg.Namespace

func (Js) installYarn() error {
	packageJSONBytes, err := ioutil.ReadFile("package.json")
	if err != nil {
		return err
	}
	var packageJSON struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err = json.Unmarshal(packageJSONBytes, &packageJSON); err != nil {
		return err
	}
	yarn, ok := packageJSON.DevDependencies["yarn"]
	if ok {
		yarn = "yarn@" + yarn
	} else {
		yarn = "yarn"
	}
	if mg.Verbose() {
		fmt.Printf("Installing Yarn %s\n", yarn)
	}
	return sh.RunV("npm", "install", "--no-package-lock", "--no-save", "--production=false", yarn)
}

func (js Js) yarn() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("yarn")); os.IsNotExist(err) {
		if err = js.installYarn(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		return sh.RunV(nodeBin("yarn"), args...)
	}, nil
}

func (js Js) webpack() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("webpack")); os.IsNotExist(err) {
		if err = js.DevDeps(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		return sh.RunV(nodeBin("webpack"), args...)
	}, nil
}

func (js Js) node() (func(args ...string) error, error) {
	return func(args ...string) error {
		return sh.RunV("node", args...)
	}, nil
}

func (js Js) babel() (func(args ...string) error, error) {
	if _, err := os.Stat(nodeBin("babel")); os.IsNotExist(err) {
		if err = js.DevDeps(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) error {
		return sh.Run(nodeBin("babel"), args...)
	}, nil
}

// DevDeps installs the javascript development dependencies.
func (js Js) DevDeps() error {
	_, err := js.yarn()
	return err
}

// Deps installs the javascript dependencies.
func (js Js) Deps() error {
	if mg.Verbose() {
		fmt.Println("Installing JS dependencies")
	}
	yarn, err := js.yarn()
	if err != nil {
		return err
	}
	return yarn("install", "--no-progress", "--production=false")
}

// BuildMain runs the webpack command with the project config.
func (js Js) BuildMain() error {
	if mg.Verbose() {
		fmt.Println("Running Webpack")
	}
	webpack, err := js.webpack()
	if err != nil {
		return err
	}
	return webpack("--config", "config/webpack.config.babel.js")
}

// BuildDll runs the webpack command with the dll config.
func (js Js) BuildDll() error {
	if mg.Verbose() {
		fmt.Println("Running Webpack")
	}
	webpack, err := js.webpack()
	if err != nil {
		return err
	}
	return webpack("--config", "config/webpack.config.dll.babel.js")
}

// Messages extracts the frontend messages via babel
func (js Js) Messages() error {
	if mg.Verbose() {
		fmt.Println("Extracting frontend messages")
	}
	babel, err := js.babel()
	if err != nil {
		return err
	}
	sh.Rm(".cache/messages")
	sh.Run("mdir", "-p", "pkg/webui/locales")
	return babel("-q", "pkg/webui")
}

// Translations builds the frontend locale files
func (js Js) Translations() error {
	node, err := js.node()
	if err != nil {
		return err
	}
	return node(".mage/translations.js")
}

// BackendTranslations builds the backend locale files
func (js Js) BackendTranslations() error {
	node, err := js.node()
	if err != nil {
		return err
	}

	return node(".mage/translations.js", "--backend-messages", "config/messages.json", "--locales", "pkg/webui/locales/.backend", "--backend-only")
}
