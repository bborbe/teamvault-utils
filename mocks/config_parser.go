// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	teamvault "github.com/bborbe/teamvault-utils"
)

type ConfigParser struct {
	ParseStub        func([]byte) ([]byte, error)
	parseMutex       sync.RWMutex
	parseArgsForCall []struct {
		arg1 []byte
	}
	parseReturns struct {
		result1 []byte
		result2 error
	}
	parseReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ConfigParser) Parse(arg1 []byte) ([]byte, error) {
	var arg1Copy []byte
	if arg1 != nil {
		arg1Copy = make([]byte, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.parseMutex.Lock()
	ret, specificReturn := fake.parseReturnsOnCall[len(fake.parseArgsForCall)]
	fake.parseArgsForCall = append(fake.parseArgsForCall, struct {
		arg1 []byte
	}{arg1Copy})
	fake.recordInvocation("Parse", []interface{}{arg1Copy})
	fake.parseMutex.Unlock()
	if fake.ParseStub != nil {
		return fake.ParseStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.parseReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ConfigParser) ParseCallCount() int {
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	return len(fake.parseArgsForCall)
}

func (fake *ConfigParser) ParseCalls(stub func([]byte) ([]byte, error)) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = stub
}

func (fake *ConfigParser) ParseArgsForCall(i int) []byte {
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	argsForCall := fake.parseArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ConfigParser) ParseReturns(result1 []byte, result2 error) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = nil
	fake.parseReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *ConfigParser) ParseReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = nil
	if fake.parseReturnsOnCall == nil {
		fake.parseReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.parseReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *ConfigParser) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ConfigParser) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ teamvault.ConfigParser = new(ConfigParser)
