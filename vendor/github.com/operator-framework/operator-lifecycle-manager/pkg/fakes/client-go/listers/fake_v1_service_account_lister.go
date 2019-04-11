// Code generated by counterfeiter. DO NOT EDIT.
package listers

import (
	sync "sync"

	v1 "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	v1a "k8s.io/client-go/listers/core/v1"
)

type FakeServiceAccountLister struct {
	ListStub        func(labels.Selector) ([]*v1.ServiceAccount, error)
	listMutex       sync.RWMutex
	listArgsForCall []struct {
		arg1 labels.Selector
	}
	listReturns struct {
		result1 []*v1.ServiceAccount
		result2 error
	}
	listReturnsOnCall map[int]struct {
		result1 []*v1.ServiceAccount
		result2 error
	}
	ServiceAccountsStub        func(string) v1a.ServiceAccountNamespaceLister
	serviceAccountsMutex       sync.RWMutex
	serviceAccountsArgsForCall []struct {
		arg1 string
	}
	serviceAccountsReturns struct {
		result1 v1a.ServiceAccountNamespaceLister
	}
	serviceAccountsReturnsOnCall map[int]struct {
		result1 v1a.ServiceAccountNamespaceLister
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeServiceAccountLister) List(arg1 labels.Selector) ([]*v1.ServiceAccount, error) {
	fake.listMutex.Lock()
	ret, specificReturn := fake.listReturnsOnCall[len(fake.listArgsForCall)]
	fake.listArgsForCall = append(fake.listArgsForCall, struct {
		arg1 labels.Selector
	}{arg1})
	fake.recordInvocation("List", []interface{}{arg1})
	fake.listMutex.Unlock()
	if fake.ListStub != nil {
		return fake.ListStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.listReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceAccountLister) ListCallCount() int {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	return len(fake.listArgsForCall)
}

func (fake *FakeServiceAccountLister) ListCalls(stub func(labels.Selector) ([]*v1.ServiceAccount, error)) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = stub
}

func (fake *FakeServiceAccountLister) ListArgsForCall(i int) labels.Selector {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	argsForCall := fake.listArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeServiceAccountLister) ListReturns(result1 []*v1.ServiceAccount, result2 error) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = nil
	fake.listReturns = struct {
		result1 []*v1.ServiceAccount
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceAccountLister) ListReturnsOnCall(i int, result1 []*v1.ServiceAccount, result2 error) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = nil
	if fake.listReturnsOnCall == nil {
		fake.listReturnsOnCall = make(map[int]struct {
			result1 []*v1.ServiceAccount
			result2 error
		})
	}
	fake.listReturnsOnCall[i] = struct {
		result1 []*v1.ServiceAccount
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceAccountLister) ServiceAccounts(arg1 string) v1a.ServiceAccountNamespaceLister {
	fake.serviceAccountsMutex.Lock()
	ret, specificReturn := fake.serviceAccountsReturnsOnCall[len(fake.serviceAccountsArgsForCall)]
	fake.serviceAccountsArgsForCall = append(fake.serviceAccountsArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("ServiceAccounts", []interface{}{arg1})
	fake.serviceAccountsMutex.Unlock()
	if fake.ServiceAccountsStub != nil {
		return fake.ServiceAccountsStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.serviceAccountsReturns
	return fakeReturns.result1
}

func (fake *FakeServiceAccountLister) ServiceAccountsCallCount() int {
	fake.serviceAccountsMutex.RLock()
	defer fake.serviceAccountsMutex.RUnlock()
	return len(fake.serviceAccountsArgsForCall)
}

func (fake *FakeServiceAccountLister) ServiceAccountsCalls(stub func(string) v1a.ServiceAccountNamespaceLister) {
	fake.serviceAccountsMutex.Lock()
	defer fake.serviceAccountsMutex.Unlock()
	fake.ServiceAccountsStub = stub
}

func (fake *FakeServiceAccountLister) ServiceAccountsArgsForCall(i int) string {
	fake.serviceAccountsMutex.RLock()
	defer fake.serviceAccountsMutex.RUnlock()
	argsForCall := fake.serviceAccountsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeServiceAccountLister) ServiceAccountsReturns(result1 v1a.ServiceAccountNamespaceLister) {
	fake.serviceAccountsMutex.Lock()
	defer fake.serviceAccountsMutex.Unlock()
	fake.ServiceAccountsStub = nil
	fake.serviceAccountsReturns = struct {
		result1 v1a.ServiceAccountNamespaceLister
	}{result1}
}

func (fake *FakeServiceAccountLister) ServiceAccountsReturnsOnCall(i int, result1 v1a.ServiceAccountNamespaceLister) {
	fake.serviceAccountsMutex.Lock()
	defer fake.serviceAccountsMutex.Unlock()
	fake.ServiceAccountsStub = nil
	if fake.serviceAccountsReturnsOnCall == nil {
		fake.serviceAccountsReturnsOnCall = make(map[int]struct {
			result1 v1a.ServiceAccountNamespaceLister
		})
	}
	fake.serviceAccountsReturnsOnCall[i] = struct {
		result1 v1a.ServiceAccountNamespaceLister
	}{result1}
}

func (fake *FakeServiceAccountLister) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	fake.serviceAccountsMutex.RLock()
	defer fake.serviceAccountsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeServiceAccountLister) recordInvocation(key string, args []interface{}) {
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

var _ v1a.ServiceAccountLister = new(FakeServiceAccountLister)
