package perform

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/eris-ltd/eris-compilers/definitions"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/log"
)

// Start the compile server
func StartServer(addrUnsecure, addrSecure, cert, key string) {
	log.Warn("Hello I'm the marmots' compilers server")
	config.InitErisDir()
	// Routes

	http.HandleFunc("/", CompileHandler)
	// Use SSL ?
	log.Debug(cert)
	if addrSecure != "" {
		log.Debug("Using HTTPS")
		log.WithField("=>", addrSecure).Debug("Listening on...")
		if err := http.ListenAndServeTLS(addrSecure, cert, key, nil); err != nil {
			log.Error("Cannot serve on http port: ", err)
			os.Exit(1)
		}
	}
	if addrUnsecure != "" {
		log.Debug("Using HTTP")
		log.WithField("=>", addrUnsecure).Debug("Listening on...")
		if err := http.ListenAndServe(addrUnsecure, nil); err != nil {
			log.Error("Cannot serve on http port: ", err)
			os.Exit(1)
		}
	}
}

// Main http request handler
// Read request, compile, build response object, write
func CompileHandler(w http.ResponseWriter, r *http.Request) {
	resp := compileResponse(w, r)
	if resp == nil {
		return
	}
	respJ, err := json.Marshal(resp)
	if err != nil {
		log.Errorln("failed to marshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(respJ)
}

// read in the files from the request, compile them
func compileResponse(w http.ResponseWriter, r *http.Request) *Response {
	// read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorln("err on read http request body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	// unmarshall body into req struct
	req := new(definitions.Request)
	err = json.Unmarshal(body, req)
	if err != nil {
		log.Errorln("err on json unmarshal of request", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	log.WithFields(log.Fields{
		"lang": req.Language,
		// "script": string(req.Script),
		"libs": req.Libraries,
		"incl": req.Includes,
	}).Debug("New Request")

	cached := CheckCached(req.Includes, req.Language)

	log.WithField("cached?", cached).Debug("Cached Item(s)")

	var resp *Response
	// if everything is cached, no need for request
	if cached {
		resp, err = CachedResponse(req.Includes, req.Language)
		if err != nil {
			log.Errorln("err during caching response", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
	} else {
		resp = compile(req)
		resp.CacheNewResponse(*req)
	}
	PrintResponse(*resp, false)
	return resp
}
