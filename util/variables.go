package util

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
)

func PreProcess(toProcess string, do *definitions.Do) (string, error) {
	// $block.... $account.... etc. should be caught. hell$$o should not
	// :$libAddr needs to be caught
	catchEr := regexp.MustCompile(`(^|\s|:)\$([a-zA-Z0-9_.]+)`)
	// If there's a match then run through the replacement process
	if catchEr.MatchString(toProcess) {
		log.WithField("match", toProcess).Debug("Replacement Match Found")

		// find what we need to catch.
		processedString := toProcess

		for _, jobMatch := range catchEr.FindAllStringSubmatch(toProcess, -1) {
			jobName := jobMatch[2]
			varName := "$" + jobName
			var innerVarName string
			var wantsInnerValues bool = false

			// first parse the reserved words.
			if strings.Contains(jobName, "block") {
				block, err := replaceBlockVariable(toProcess, do)
				if err != nil {
					log.WithField("err", err).Error("Error replacing block variable.")
					return "", err
				}
				/*log.WithFields(log.Fields{
					"var": toProcess,
					"res": block,
				}).Debug("Fixing Variables =>")*/
				processedString = strings.Replace(processedString, toProcess, block, 1)
			}

			if strings.Contains(jobName, ".") { //for functions with multiple returns
				wantsInnerValues = true
				var splitStr = strings.Split(jobName, ".")
				jobName = splitStr[0]
				innerVarName = splitStr[1]
			}

			// second we loop through the jobNames to do a result replace
			for _, job := range do.Package.Jobs {
				if string(jobName) == job.JobName {
					if wantsInnerValues {
						for _, innerVal := range job.JobVars {
							if innerVal.Name == innerVarName { //find the value we want from the bunch
								processedString = strings.Replace(processedString, varName, innerVal.Value, 1)
								log.WithFields(log.Fields{
									"job":     string(jobName),
									"varName": innerVarName,
									"result":  innerVal.Value,
								}).Debug("Fixing Inner Vars =>")
							}
						}
					} else {
						log.WithFields(log.Fields{
							"var": string(jobName),
							"res": job.JobResult,
						}).Debug("Fixing Variables =>")
						processedString = strings.Replace(processedString, varName, job.JobResult, 1)
					}
				}
			}
		}
		return processedString, nil
	}

	// if no matches, return original
	return toProcess, nil
}

func replaceBlockVariable(toReplace string, do *definitions.Do) (string, error) {
	log.WithFields(log.Fields{
		"chain": do.ChainName,
		"var":   toReplace,
	}).Debug("Correcting $block variable")
	blockHeight, err := GetBlockHeight(do)
	block := strconv.Itoa(blockHeight)
	log.WithField("=>", block).Debug("Current height is")
	if err != nil {
		return "", err
	}

	if toReplace == "$block" {
		log.WithField("=>", block).Debug("Replacement (=)")
		return block, nil
	}

	catchEr := regexp.MustCompile(`\$block\+(\d*)`)
	if catchEr.MatchString(toReplace) {
		height := catchEr.FindStringSubmatch(toReplace)[1]
		h1, err := strconv.Atoi(height)
		if err != nil {
			return "", err
		}
		h2, err := strconv.Atoi(block)
		if err != nil {
			return "", err
		}
		height = strconv.Itoa(h1 + h2)
		log.WithField("=>", height).Debug("Replacement (+)")
		return height, nil
	}

	catchEr = regexp.MustCompile(`\$block\-(\d*)`)
	if catchEr.MatchString(toReplace) {
		height := catchEr.FindStringSubmatch(toReplace)[1]
		h1, err := strconv.Atoi(height)
		if err != nil {
			return "", err
		}
		h2, err := strconv.Atoi(block)
		if err != nil {
			return "", err
		}
		height = strconv.Itoa(h1 - h2)
		log.WithField("=>", height).Debug("Replacement (-)")
		return height, nil
	}

	log.WithField("=>", toReplace).Debug("Replacement (unknown)")
	return toReplace, nil
}

func PreProcessInputData(function string, data interface{}, do *definitions.Do, constructor bool) (string, []string, error) {
	var callDataArray []string
	var callArray []string
	if function == "" && !constructor {
		if reflect.TypeOf(data).Kind() == reflect.Slice {
			return "", []string{""}, fmt.Errorf("Incorrect formatting of epm.yaml. Please update it to include a function field.")
		}
		function = strings.Split(data.(string), " ")[0]
		callArray = strings.Split(data.(string), " ")[1:]
		for _, val := range callArray {
			output, _ := PreProcess(val, do)
			callDataArray = append(callDataArray, output)
		}
	} else if data != nil {
		if reflect.TypeOf(data).Kind() != reflect.Slice {
			if constructor {
				log.Warn("Deprecation Warning: Your deploy job is currently using a soon to be deprecated way of declaring constructor values. Please remember to update your run file to store them as a array rather than a string. See documentation for further details.")
				callArray = strings.Split(data.(string), " ")
				for _, val := range callArray {
					output, _ := PreProcess(val, do)
					callDataArray = append(callDataArray, output)
				}
				return function, callDataArray, nil
			} else {
				return "", make([]string, 0), fmt.Errorf("Incorrect formatting of epm.yaml file. Please update it to include a function field.")
			}
		}
		val := reflect.ValueOf(data)
		for i := 0; i < val.Len(); i++ {
			s := val.Index(i)
			var newString string
			switch s.Interface().(type) {
			case bool:
				newString = strconv.FormatBool(s.Interface().(bool))
			case int, int32, int64:
				newString = strconv.FormatInt(int64(s.Interface().(int)), 10)
			case []interface{}:
				var args []string
				for _, index := range s.Interface().([]interface{}) {
					value := reflect.ValueOf(index)
					var stringified string
					switch value.Kind() {
					case reflect.Int:
						stringified = strconv.FormatInt(value.Int(), 10)
					case reflect.String:
						stringified = value.String()
					}
					index, _ = PreProcess(stringified, do)
					args = append(args, stringified)
				}
				newString = "[" + strings.Join(args, ",") + "]"
				log.Debug(newString)
			default:
				newString = s.Interface().(string)
			}
			newString, _ = PreProcess(newString, do)
			callDataArray = append(callDataArray, newString)
		}
	}
	return function, callDataArray, nil
}

func PreProcessLibs(libs string, do *definitions.Do) (string, error) {
	libraries, _ := PreProcess(libs, do)
	if libraries != "" {
		pairs := strings.Split(libraries, ",")
		libraries = strings.Join(pairs, " ")
	}
	log.WithField("=>", libraries).Debug("Library String")
	return libraries, nil
}

func GetReturnValue(vars []*definitions.Variable) string {
	var result []string

	if len(vars) > 1 {
		for _, value := range vars {
			log.WithField("=>", []byte(value.Value)).Debug("Value")
			result = append(result, value.Value)
		}
		return "(" + strings.Join(result, ", ") + ")"
	} else if len(vars) == 1 {
		log.Debug("Debugging: ", vars[0].Value)
		return vars[0].Value
	} else {
		return ""
	}
}
