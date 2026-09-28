// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: BaseMethods.BeginEnd and ParseUtil.checkMaxMacros.
package tex

import "strings"

func (p *parser) readEnvironmentName(name string) (string, error) {
	env, _, err := p.readArgument(name, false)
	if err != nil {
		return "", err
	}
	if strings.Contains(env, `\`) {
		return "", texError("InvalidEnv", "Invalid environment name '%s'", env)
	}
	return env, nil
}

func (p *parser) countEnvironment() error {
	p.state.macroCount++
	if p.state.macroCount > maxMacros {
		return texError("MaxMacroSub2", "MathJax maximum substitution count exceeded; is there a recursive latex environment?")
	}
	return nil
}

// BeginEnd returns an EndItem before checkMaxMacros only for a registered
// environment with a falsy first macro argument. Truthy entries (including
// empheq) and user-defined ends take the charged environment-dispatch path.
// Retained maps keep this independent of Go's supported environment handlers.
func (p *parser) sourceEndReturnsItem(name string) bool {
	if _, defined := p.state.environments[name]; defined {
		return false
	}
	for _, table := range mjSourceMaps {
		if table.Kind != mjSourceEnvironmentMap || (!p.state.augmentedPackages && strings.Contains(table.Source, "/empheq/")) {
			continue
		}
		for _, entry := range table.Entries {
			if entry.Name == name {
				args, _ := entry.Value.(mjSourceList)
				if len(args) < 2 {
					return true
				}
				_, undefined := args[1].(mjSourceUndefined)
				return undefined || !limitsTruthy(args[1])
			}
		}
	}
	return false
}
