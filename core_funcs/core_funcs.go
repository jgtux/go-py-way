package core_funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

func PyWayRecipe(recipe string, keys map[string]interface{}, mutableKeys []string) error {
	for _, k := range mutableKeys {
		if _, ok := keys[k]; !ok {
			return fmt.Errorf("ERROR: unknown mutable key '%s' in mutable_keys slice", k)
		}
	}

	// check if keys are being used in "recipe"
	if !checkVariablesInPythonFile(recipe, keys) {
		return fmt.Errorf("ERROR: Not all keys are being used in the recipe")
	}

	// clean "recipe" indentation and inject keys in into "recipe" 
	indentationCleanRec := cleanPythonIndentation(recipe) 
	script := injectVariablesIntoScript(indentationCleanRec, keys, mutableKeys)

	// execute script and captures json exit 
	out, err := executePythonScript(script)
	if err != nil {
		return fmt.Errorf("ERROR: Failed to execute recipe: %v", err)
	}

	// unmarshal json into keys
	sanitizeStdout(&out)
	if err != nil {
		return fmt.Errorf("ERROR: Failed to sanitize output: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return fmt.Errorf("ERROR: Failed to parse output JSON: %v", err)
	}

	for _, k := range mutableKeys {
		if val, ok := result[k]; ok {
			keys[k] = val
		}
	}

	return nil
}

func checkVariablesInPythonFile(recipe string, keys map[string]interface{}) bool {
	for variable := range keys {
		if !isVariableUsed(recipe, variable) {
			return false
		}
	}
	return true
}

func isVariableUsed(content, variable string) bool {
	// regular expression for checking words in sentence
	pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(variable))
	re := regexp.MustCompile(pattern)

	cleanContent := removeCommentsAndStrings(content)


	return re.MatchString(cleanContent)
}

func removeCommentsAndStrings(content string) string {
	// remove triple strings (""", ''')
	tripleDouble := regexp.MustCompile(`"""[\s\S]*?"""`)
	content = tripleDouble.ReplaceAllString(content, "")
	tripleSingle := regexp.MustCompile(`'''[\s\S]*?'''`)
	content = tripleSingle.ReplaceAllString(content, "")

	// remove simple string ("", '')
	doubleQuoted := regexp.MustCompile(`"([^"\\]|\\.)*"`)
	content = doubleQuoted.ReplaceAllString(content, "")
	singleQuoted := regexp.MustCompile(`'([^'\\]|\\.)*'`)
	content = singleQuoted.ReplaceAllString(content, "")

	// remove '#' comments
	comments := regexp.MustCompile(`#.*$`)
	content = comments.ReplaceAllString(content, "")

	return content
}

func cleanPythonIndentation(code string) string {
	lines := strings.Split(code, "\n")

	minIndent := -1
	for _, line := range lines {
		trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
		if trimmed == "" {
			continue
		}
		indent := len(line) - len(trimmed)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent > 0 {
		for i, line := range lines {
			if len(line) >= minIndent {
				lines[i] = line[minIndent:]
			}
		}
	}

	return strings.Join(lines, "\n")
}

func injectVariablesIntoScript(recipe string, keys map[string]interface{}, mutableKeys []string) string {
	var sb strings.Builder

	sb.WriteString("import json\n")

	// Declara variáveis em Python
	for k, v := range keys {
		sb.WriteString(fmt.Sprintf("%s = %s\n", k, toPythonValue(v)))
	}

	sb.WriteString("\n")
	sb.WriteString(recipe)
	sb.WriteString("\n\n")

	// needs to improve to use be able to print other values  
	sb.WriteString("print(json.dumps({\n")
	for i, k := range mutableKeys {
		comma := ","
		if i == len(mutableKeys)-1 {
			comma = ""
		}
		sb.WriteString(fmt.Sprintf("  \"%s\": %s%s\n", k, k, comma))
	}
	sb.WriteString("}))\n")

	return sb.String()
}

// converts values to python  
func toPythonValue(val interface{}) string {
	if val == nil {
		return "None"
	}

	switch v := val.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case bool:
		if v {
			return "True"
		}
		return "False"
	case float64, float32, int, int64, int32:
		return fmt.Sprintf("%v", v)
	default:
		// json for objects 
		jsonVal, _ := json.Marshal(v)
		return string(jsonVal)
	}
}

func executePythonScript(pythonCode string) (string, error) {
	cmd := exec.Command("python3", "-c", pythonCode)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Python execution error: %v\n%s", err, stderr.String())
	}

	return strings.TrimSpace(out.String()), nil
}

func sanitizeStdout(output *string) error {
	re := regexp.MustCompile(`(?m)^\s*[\[{].*[\]}]\s*$`)

	matches := re.FindAllString(*output, -1)
	if len(matches) > 0 {
		*output = matches[len(matches)-1]
	} else {
		*output = "" 
	}
	return nil
}
