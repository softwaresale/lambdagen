package model

import (
	"regexp"
	"strings"
)

// ObjectRole describes the different annotations/tags available to lambdagen
type ObjectRole struct {
	Type string
	Args string
}

const (
	ObjectRoleServiceTp   = "service"      // used on structs, defines that this object is a service
	ObjectRoleServiceInit = "service_init" // function that initializes our service
	ObjectRoleHandlerTp   = "handler"      // this function belongs to a service, handles an endpoint
	ObjectRolePathVar     = "pathvar"      // this field is a path variable
	ObjectRoleQueryParam  = "queryvar"     // this field is a query variable
	ObjectRoleBody        = "body"         // this field is the request body
)

func IsValidRoleStr(roleStr string) bool {
	switch roleStr {
	case ObjectRoleServiceTp, ObjectRoleServiceInit, ObjectRoleHandlerTp, ObjectRolePathVar, ObjectRoleQueryParam, ObjectRoleBody:
		return true
	default:
		return false
	}
}

func (role ObjectRole) IsValid() bool {
	return IsValidRoleStr(role.Type)
}

func (role ObjectRole) GetServiceConfig() map[string]string {

	// return nil if incorrect role
	if role.Type != ObjectRoleServiceTp {
		return nil
	}

	parser := regexp.MustCompile(`([a-zA-Z_]\w*)=(\S+)`)

	config := make(map[string]string)
	matches := parser.FindAllStringSubmatch(role.Args, -1)
	for _, match := range matches {
		config[match[1]] = match[2]
	}

	return config
}

// ParseObjectRoleDocstring parses a docstring looking for a valid object role. It finds the first role present. If
// a role is present, then return the role and true. Otherwise, return empty and false.
func ParseObjectRoleDocstring(docsOrTags string) (ObjectRole, bool) {
	roleExtractor := regexp.MustCompile(`lambdagen:(\S+)`)

	fields := strings.Fields(docsOrTags)

	roleIdx := -1
	role := ""
	for idx, field := range fields {
		matches := roleExtractor.FindStringSubmatch(field)
		if matches == nil {
			continue
		}

		if IsValidRoleStr(matches[1]) {
			roleIdx = idx
			role = matches[1]
			break
		}
	}

	if roleIdx == -1 {
		return ObjectRole{}, false
	}

	// find the "::" separator after the role
	separatorIdx := -1
	for idx, field := range fields[roleIdx:] {
		if field == "::" {
			separatorIdx = idx
			break
		}
	}

	// if we found args, pull them
	argStr := ""
	if separatorIdx != -1 {
		argStr = strings.Join(fields[separatorIdx+1:], " ")
	}

	return ObjectRole{
		Type: role,
		Args: argStr,
	}, true
}

func ParseObjectRoleTag(tag string) (ObjectRole, bool) {
	parser := regexp.MustCompile(`lambdagen:"([^)]+)"`)

	fields := strings.Fields(tag)
	for _, field := range fields {
		matches := parser.FindStringSubmatch(field)
		if matches == nil {
			continue
		}

		params := strings.Split(matches[1], ",")
		if len(params) == 0 {
			return ObjectRole{}, false
		}

		// param 1 should be a valid role
		if !IsValidRoleStr(params[0]) {
			return ObjectRole{}, false
		}

		return ObjectRole{
			Type: params[0],
			Args: strings.Join(params[1:], ","),
		}, true
	}

	return ObjectRole{}, false
}
