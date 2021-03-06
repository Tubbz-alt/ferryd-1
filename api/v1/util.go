//
// Copyright © 2017-2020 Solus Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1

import (
	"errors"
	"fmt"
	log "github.com/DataDrake/waterlog"
	"github.com/valyala/fasthttp"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
)

// getMethodOrigin helps us determine the caller so that we can print
// an appropriate method name into the log without tons of boilerplate
func getMethodCaller() string {
	n, _, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	if details := runtime.FuncForPC(n); details != nil {
		return details.Name()
	}
	return ""
}

func formURI(part string) string {
	return fmt.Sprintf("http://localhost.localdomain:0/%s", part)
}

func readError(in io.Reader) error {
	raw, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	return errors.New(string(raw))
}

func writeErrorString(ctx *fasthttp.RequestCtx, e string, code int) {
	log.Errorln(e)
	ctx.Error(e, code)
}

func writeError(ctx *fasthttp.RequestCtx, err error, code int) {
	writeErrorString(ctx, err.Error(), code)
}

func readID(resp *http.Response) (id int, err error) {
	// Check for failure
	if resp.StatusCode != http.StatusOK {
		err = readError(resp.Body)
		return
	}
	// Read body as ID
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	id, err = strconv.Atoi(string(raw))
	return
}

func writeID(ctx *fasthttp.RequestCtx, id int) {
	s := strconv.Itoa(id)
	ctx.SetBodyString(s)
	return
}
