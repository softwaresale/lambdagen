package model

import "strings"

// ObjectRole describes the different annotations/tags available to lambdagen
type ObjectRole struct {
	Type string
	Args string
}

const (
	ObjectRoleServiceTp   = "lambdagen:service"      // used on structs, defines that this object is a service
	ObjectRoleServiceInit = "lambdagen:service_init" // function that initializes our service
	ObjectRoleHandlerTp   = "lambdagen:handler"      // this function belongs to a service, handles an endpoint
	ObjectRolePathVar     = "lambdagen:pathvar"      // this field is a path variable
	ObjectRoleQueryParam  = "lambdagen:queryvar"     // this field is a query variable
	ObjectRoleBody        = "lambdagen:body"         // this field is the request body
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

// ParseObjectRoleDocstring parses a docstring looking for a valid object role. It finds the first role present. If
// a role is present, then return the role and true. Otherwise, return empty and false.
func ParseObjectRoleDocstring(docstrings ...string) (ObjectRole, bool) {
	for _, docstring := range docstrings {
		fields := strings.Fields(docstring)

		roleIdx := -1
		for idx, field := range fields {
			if IsValidRoleStr(field) {
				roleIdx = idx
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
			Type: fields[roleIdx],
			Args: argStr,
		}, true
	}

	return ObjectRole{}, false
}
