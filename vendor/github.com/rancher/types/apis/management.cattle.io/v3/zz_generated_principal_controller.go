package v3

import (
	"context"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/controller"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	PrincipalGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "Principal",
	}
	PrincipalResource = metav1.APIResource{
		Name:         "principals",
		SingularName: "principal",
		Namespaced:   false,
		Kind:         PrincipalGroupVersionKind.Kind,
	}
)

type PrincipalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Principal
}

type PrincipalHandlerFunc func(key string, obj *Principal) error

type PrincipalLister interface {
	List(namespace string, selector labels.Selector) (ret []*Principal, err error)
	Get(namespace, name string) (*Principal, error)
}

type PrincipalController interface {
	Informer() cache.SharedIndexInformer
	Lister() PrincipalLister
	AddHandler(handler PrincipalHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type PrincipalInterface interface {
	ObjectClient() *clientbase.ObjectClient
	Create(*Principal) (*Principal, error)
	Get(name string, opts metav1.GetOptions) (*Principal, error)
	Update(*Principal) (*Principal, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*PrincipalList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() PrincipalController
	AddSyncHandler(sync PrincipalHandlerFunc)
	AddLifecycle(name string, lifecycle PrincipalLifecycle)
}

type principalLister struct {
	controller *principalController
}

func (l *principalLister) List(namespace string, selector labels.Selector) (ret []*Principal, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*Principal))
	})
	return
}

func (l *principalLister) Get(namespace, name string) (*Principal, error) {
	var key string
	if namespace != "" {
		key = namespace + "/" + name
	} else {
		key = name
	}
	obj, exists, err := l.controller.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    PrincipalGroupVersionKind.Group,
			Resource: "principal",
		}, name)
	}
	return obj.(*Principal), nil
}

type principalController struct {
	controller.GenericController
}

func (c *principalController) Lister() PrincipalLister {
	return &principalLister{
		controller: c,
	}
}

func (c *principalController) AddHandler(handler PrincipalHandlerFunc) {
	c.GenericController.AddHandler(func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*Principal))
	})
}

type principalFactory struct {
}

func (c principalFactory) Object() runtime.Object {
	return &Principal{}
}

func (c principalFactory) List() runtime.Object {
	return &PrincipalList{}
}

func (s *principalClient) Controller() PrincipalController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.principalControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(PrincipalGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &principalController{
		GenericController: genericController,
	}

	s.client.principalControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type principalClient struct {
	client       *Client
	ns           string
	objectClient *clientbase.ObjectClient
	controller   PrincipalController
}

func (s *principalClient) ObjectClient() *clientbase.ObjectClient {
	return s.objectClient
}

func (s *principalClient) Create(o *Principal) (*Principal, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*Principal), err
}

func (s *principalClient) Get(name string, opts metav1.GetOptions) (*Principal, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*Principal), err
}

func (s *principalClient) Update(o *Principal) (*Principal, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*Principal), err
}

func (s *principalClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *principalClient) List(opts metav1.ListOptions) (*PrincipalList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*PrincipalList), err
}

func (s *principalClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

func (s *principalClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *principalClient) AddSyncHandler(sync PrincipalHandlerFunc) {
	s.Controller().AddHandler(sync)
}

func (s *principalClient) AddLifecycle(name string, lifecycle PrincipalLifecycle) {
	sync := NewPrincipalLifecycleAdapter(name, s, lifecycle)
	s.AddSyncHandler(sync)
}
