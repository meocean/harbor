/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	cfg "github.com/vmware/harbor/src/adminserver/systemcfg"
	"github.com/vmware/harbor/src/common/utils/log"
)

func isAuthenticated(r *http.Request) (bool, error) {
	uiSecret := os.Getenv("UI_SECRET")
	jobserviceSecret := os.Getenv("JOBSERVICE_SECRET")
	c, err := r.Cookie("secret")
	if err != nil {
		if err == http.ErrNoCookie {
			return false, nil
		}
		return false, err
	}
	return c != nil && (c.Value == uiSecret ||
		c.Value == jobserviceSecret), nil
}

// ListCfgs lists configurations
func ListCfgs(w http.ResponseWriter, r *http.Request) {
	authenticated, err := isAuthenticated(r)
	if err != nil {
		log.Errorf("failed to check whether the request is authenticated or not: %v", err)
		handleInternalServerError(w)
		return
	}

	if !authenticated {
		handleUnauthorized(w)
		return
	}

	cfg, err := cfg.GetSystemCfg()
	if err != nil {
		log.Errorf("failed to get system configurations: %v", err)
		handleInternalServerError(w)
		return
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Errorf("failed to marshal configurations: %v", err)
		handleInternalServerError(w)
		return
	}
	if _, err = w.Write(b); err != nil {
		log.Errorf("failed to write response: %v", err)
	}
}

// UpdateCfgs updates configurations
func UpdateCfgs(w http.ResponseWriter, r *http.Request) {
	authenticated, err := isAuthenticated(r)
	if err != nil {
		log.Errorf("failed to check whether the request is authenticated or not: %v", err)
		handleInternalServerError(w)
		return
	}

	if !authenticated {
		handleUnauthorized(w)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read request body: %v", err)
		handleInternalServerError(w)
		return
	}

	m := map[string]interface{}{}
	if err = json.Unmarshal(b, &m); err != nil {
		handleBadRequestError(w, err.Error())
		return
	}

	if err = cfg.UpdateSystemCfg(m); err != nil {
		log.Errorf("failed to update system configurations: %v", err)
		handleInternalServerError(w)
		return
	}
}
