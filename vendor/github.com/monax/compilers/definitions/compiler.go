package definitions

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"

	"github.com/monax/monax/log"
)

type Compiler struct {
	Config LangConfig
	Lang   string
}

// New Request object from script and map of include files
func (c *Compiler) CompilerRequest(file string, includes map[string]*IncludedFiles, libs string, optimize bool, hashFileReplacement map[string]string) *Request {
	if includes == nil {
		includes = make(map[string]*IncludedFiles)
	}
	return &Request{
		Language:        c.Lang,
		Includes:        includes,
		Libraries:       libs,
		Optimize:        optimize,
		FileReplacement: hashFileReplacement,
	}
}

// Find all matches to the include regex
// Replace filenames with hashes
func (c *Compiler) ReplaceIncludes(code []byte, dir, file string, includes map[string]*IncludedFiles, hashFileReplacement map[string]string) ([]byte, error) {
	// find includes, load those as well
	regexPattern := c.IncludeRegex()
	var regExpression *regexp.Regexp
	var err error
	if regExpression, err = regexp.Compile(regexPattern); err != nil {
		return nil, err
	}
	OriginObjectNames, err := c.extractObjectNames(code)
	if err != nil {
		return nil, err
	}
	// replace all includes with hash of included imports
	// make sure to return hashes of includes so we can cache check them too
	// do it recursively
	code = regExpression.ReplaceAllFunc(code, func(s []byte) []byte {
		log.WithField("=>", string(s)).Debug("Include Replacer result")
		s, err := c.includeReplacer(regExpression, s, dir, includes, hashFileReplacement)
		if err != nil {
			log.Error("ERR!:", err)
		}
		return s
	})

	originHash := sha256.Sum256(code)
	origin := hex.EncodeToString(originHash[:])
	origin += "." + c.Lang

	includeFile := &IncludedFiles{
		ObjectNames: OriginObjectNames,
		Script:      code,
	}

	includes[origin] = includeFile
	hashFileReplacement[origin] = file

	return code, nil
}

// read the included file, hash it; if we already have it, return include replacement
// if we don't, run replaceIncludes on it (recursive)
// modifies the "includes" map
func (c *Compiler) includeReplacer(r *regexp.Regexp, originCode []byte, dir string, included map[string]*IncludedFiles, hashFileReplacement map[string]string) ([]byte, error) {
	// regex look for strings that would match the import statement
	m := r.FindStringSubmatch(string(originCode))
	match := m[3]
	log.WithField("=>", match).Debug("Match")
	// load the file
	newFilePath := path.Join(dir, match)
	incl_code, err := ioutil.ReadFile(newFilePath)
	if err != nil {
		log.Errorln("failed to read include file", err)
		return nil, fmt.Errorf("Failed to read include file: %s", err.Error())
	}

	// take hash before replacing includes to see if we've already parsed this file
	hash := sha256.Sum256(incl_code)
	includeHash := hex.EncodeToString(hash[:])
	log.WithField("=>", includeHash).Debug("Included Code's Hash")
	if _, ok := included[includeHash]; ok {
		//then replace
		fullReplacement := strings.SplitAfter(m[0], m[2])
		fullReplacement[1] = includeHash + "." + c.Lang + "\""
		ret := strings.Join(fullReplacement, "")
		return []byte(ret), nil
	}

	// recursively replace the includes for this file
	this_dir := path.Dir(newFilePath)
	incl_code, err = c.ReplaceIncludes(incl_code, this_dir, newFilePath, included, hashFileReplacement)
	if err != nil {
		return nil, err
	}

	// compute hash
	hash = sha256.Sum256(incl_code)
	h := hex.EncodeToString(hash[:])

	//Starting with full regex string,
	//Split strings from the quotation mark and then,
	//assuming 3 array cells, replace the middle one.
	fullReplacement := strings.SplitAfter(m[0], m[2])
	fullReplacement[1] = h + "." + c.Lang + m[4]
	ret := []byte(strings.Join(fullReplacement, ""))
	return ret, nil
}

// Return the regex string to match include statements
func (c *Compiler) IncludeRegex() string {
	return c.Config.IncludeRegex
}

func (c *Compiler) extractObjectNames(script []byte) ([]string, error) {
	regExpression, err := regexp.Compile("(contract|library) (.+?) (is)?(.+?)?({)")
	if err != nil {
		return nil, err
	}
	objectNamesList := regExpression.FindAllSubmatch(script, -1)
	var objects []string
	for _, objectNames := range objectNamesList {
		objects = append(objects, string(objectNames[2]))
	}
	return objects, nil
}
