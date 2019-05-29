/*
 * Copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package filter

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Operation int

const (
	UK Operation = iota
	EQ
	NE
	GT
	LT
	GE
	LE
)

func toOp(op string) Operation {
	switch op {
	case "=":
		return EQ
	case "!=":
		return NE
	case ">":
		return GT
	case "<":
		return LT
	case ">=":
		return GE
	case "<=":
		return LE
	default:
		return UK
	}
}

type FilterTerm struct {
	Op    Operation
	Value string
}

type Filter map[string]FilterTerm

var termRE = regexp.MustCompile("^\\s*([a-zA-Z_][.a-zA-Z0-9_]*)\\s*(<=|>=|<|>|!=|=)\\s*(.+)\\s*$")

// Parse parses a comma separated list of filter terms
func Parse(spec string) (Filter, error) {
	filter := make(map[string]FilterTerm)
	terms := strings.Split(spec, ",")

	// Each term is in the form <key><op><value>
	for _, term := range terms {
		parts := termRE.FindAllStringSubmatch(term, -1)
		if parts == nil {
			return nil, fmt.Errorf("Unable to parse filter term '%s'", term)
		}
		filter[parts[0][1]] = FilterTerm{
			Op:    toOp(parts[0][2]),
			Value: parts[0][3],
		}
	}
	return filter, nil
}

func (f Filter) Process(data interface{}) (interface{}, error) {
	var list []interface{}
	var islist bool
	if list, islist = data.([]interface{}); !islist {
		list = []interface{}{data}
	}

	// Now we have the data as a list, so we can processing it
	// identically
	var result []interface{}
	for _, item := range list {
		if f.Evaluate(item) {
			result = append(result, item)
		}
	}

	// If we were given a single item, make sure we return a single
	// item or nil
	if !islist {
		if len(result) == 0 {
			return nil, nil
		}
		return result[0], nil
	}

	return result, nil
}

func (f Filter) Evaluate(item interface{}) bool {
	val := reflect.ValueOf(item)

	for k, v := range f {
		field := val.FieldByName(k)
		if !field.IsValid() {
			return false
		}

		switch v.Op {
		case EQ:
			if field.String() != v.Value {
				return false
			}
		default:
			// For unsupported operations, always pass
		}
	}
	return true
}
