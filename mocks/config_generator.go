// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"context"
	"sync"

	teamvault "github.com/bborbe/teamvault-utils"
)

type ConfigGenerator struct {
	GenerateStub        func(context.Context, teamvault.SourceDirectory, teamvault.TargetDirectory) error
	generateMutex       sync.RWMutex
	generateArgsForCall []struct {
		arg1 context.Context
		arg2 teamvault.SourceDirectory
		arg3 teamvault.TargetDirectory
	}
	generateReturns struct {
		result1 error
	}
	generateReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ConfigGenerator) Generate(arg1 context.Context, arg2 teamvault.SourceDirectory, arg3 teamvault.TargetDirectory) error {
	fake.generateMutex.Lock()
	ret, specificReturn := fake.generateReturnsOnCall[len(fake.generateArgsForCall)]
	fake.generateArgsForCall = append(fake.generateArgsForCall, struct {
		arg1 context.Context
		arg2 teamvault.SourceDirectory
		arg3 teamvault.TargetDirectory
	}{arg1, arg2, arg3})
	stub := fake.GenerateStub
	fakeReturns := fake.generateReturns
	fake.recordInvocation("Generate", []interface{}{arg1, arg2, arg3})
	fake.generateMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *ConfigGenerator) GenerateCallCount() int {
	fake.generateMutex.RLock()
	defer fake.generateMutex.RUnlock()
	return len(fake.generateArgsForCall)
}

func (fake *ConfigGenerator) GenerateCalls(stub func(context.Context, teamvault.SourceDirectory, teamvault.TargetDirectory) error) {
	fake.generateMutex.Lock()
	defer fake.generateMutex.Unlock()
	fake.GenerateStub = stub
}

func (fake *ConfigGenerator) GenerateArgsForCall(i int) (context.Context, teamvault.SourceDirectory, teamvault.TargetDirectory) {
	fake.generateMutex.RLock()
	defer fake.generateMutex.RUnlock()
	argsForCall := fake.generateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *ConfigGenerator) GenerateReturns(result1 error) {
	fake.generateMutex.Lock()
	defer fake.generateMutex.Unlock()
	fake.GenerateStub = nil
	fake.generateReturns = struct {
		result1 error
	}{result1}
}

func (fake *ConfigGenerator) GenerateReturnsOnCall(i int, result1 error) {
	fake.generateMutex.Lock()
	defer fake.generateMutex.Unlock()
	fake.GenerateStub = nil
	if fake.generateReturnsOnCall == nil {
		fake.generateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.generateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ConfigGenerator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.generateMutex.RLock()
	defer fake.generateMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ConfigGenerator) recordInvocation(key string, args []interface{}) {
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

var _ teamvault.ConfigGenerator = new(ConfigGenerator)
