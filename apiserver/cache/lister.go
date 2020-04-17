package cache

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	applicationv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	applicationgroupv1 "github.com/chinamobile/nlpt/crds/applicationgroup/api/v1"
	applyv1 "github.com/chinamobile/nlpt/crds/apply/api/v1"
	clientauthv1 "github.com/chinamobile/nlpt/crds/clientauth/api/v1"
	dataservicev1 "github.com/chinamobile/nlpt/crds/dataservice/api/v1"
	datasourcev1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	restrictionv1 "github.com/chinamobile/nlpt/crds/restriction/api/v1"
	serviceunitv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	serviceunitgroupv1 "github.com/chinamobile/nlpt/crds/serviceunitgroup/api/v1"
	topicv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	topicgroupv1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	trafficcontrolv1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
)

type ApiLister struct {
	l *typedLister
}

func (c *Listers) ApiLister() *ApiLister {
	lister, err := c.WithResource("apis")
	if err != nil {
		panic(err)
	}
	return &ApiLister{
		lister,
	}
}

func (a *ApiLister) Get(namespace, name string) (*apiv1.Api, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*apiv1.Api); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to apiv1.Api", o)
}

func (a *ApiLister) List(namespace string, lo metav1.ListOptions) ([]*apiv1.Api, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*apiv1.Api, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*apiv1.Api)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to apiv1.Api", o)
		}
	}
	return result, nil
}

type ApplicationLister struct {
	l *typedLister
}

func (c *Listers) ApplicationLister() *ApplicationLister {
	lister, err := c.WithResource("applications")
	if err != nil {
		panic(err)
	}
	return &ApplicationLister{
		lister,
	}
}

func (a *ApplicationLister) Get(namespace, name string) (*applicationv1.Application, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*applicationv1.Application); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to applicationv1.Application", o)
}

func (a *ApplicationLister) List(namespace string, lo metav1.ListOptions) ([]*applicationv1.Application, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*applicationv1.Application, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*applicationv1.Application)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to applicationv1.Application", o)
		}
	}
	return result, nil
}

type ApplicationGroupLister struct {
	l *typedLister
}

func (c *Listers) ApplicationGroupLister() *ApplicationGroupLister {
	lister, err := c.WithResource("applicationgroups")
	if err != nil {
		panic(err)
	}
	return &ApplicationGroupLister{
		lister,
	}
}

func (a *ApplicationGroupLister) Get(namespace, name string) (*applicationgroupv1.ApplicationGroup, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*applicationgroupv1.ApplicationGroup); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to applicationgroupv1.ApplicationGroup", o)
}

func (a *ApplicationGroupLister) List(namespace string, lo metav1.ListOptions) ([]*applicationgroupv1.ApplicationGroup, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*applicationgroupv1.ApplicationGroup, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*applicationgroupv1.ApplicationGroup)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to applicationgroupv1.ApplicationGroup", o)
		}
	}
	return result, nil
}

type ApplyLister struct {
	l *typedLister
}

func (c *Listers) ApplyLister() *ApplyLister {
	lister, err := c.WithResource("applies")
	if err != nil {
		panic(err)
	}
	return &ApplyLister{
		lister,
	}
}

func (a *ApplyLister) Get(namespace, name string) (*applyv1.Apply, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*applyv1.Apply); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to applyv1.Apply", o)
}

func (a *ApplyLister) List(namespace string, lo metav1.ListOptions) ([]*applyv1.Apply, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*applyv1.Apply, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*applyv1.Apply)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to applyv1.Apply", o)
		}
	}
	return result, nil
}

type ClientauthLister struct {
	l *typedLister
}

func (c *Listers) ClientauthLister() *ClientauthLister {
	lister, err := c.WithResource("clientauths")
	if err != nil {
		panic(err)
	}
	return &ClientauthLister{
		lister,
	}
}

func (a *ClientauthLister) Get(namespace, name string) (*clientauthv1.Clientauth, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*clientauthv1.Clientauth); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to clientauthv1.Clientauth", o)
}

func (a *ClientauthLister) List(namespace string, lo metav1.ListOptions) ([]*clientauthv1.Clientauth, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*clientauthv1.Clientauth, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*clientauthv1.Clientauth)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to clientauthv1.Clientauth", o)
		}
	}
	return result, nil
}

type DataserviceListerer struct {
	l *typedLister
}

func (c *Listers) DataserviceLister() *DataserviceListerer {
	lister, err := c.WithResource("dataservices")
	if err != nil {
		panic(err)
	}
	return &DataserviceListerer{
		lister,
	}
}

func (a *DataserviceListerer) Get(namespace, name string) (*dataservicev1.Dataservice, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*dataservicev1.Dataservice); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to dataservicev1.Dataservice", o)
}

func (a *DataserviceListerer) List(namespace string, lo metav1.ListOptions) ([]*dataservicev1.Dataservice, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*dataservicev1.Dataservice, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*dataservicev1.Dataservice)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to dataservicev1.Dataservice", o)
		}
	}
	return result, nil
}

type DatasourceLister struct {
	l *typedLister
}

func (c *Listers) DatasourceLister() *DatasourceLister {
	lister, err := c.WithResource("datasources")
	if err != nil {
		panic(err)
	}
	return &DatasourceLister{
		lister,
	}
}

func (a *DatasourceLister) Get(namespace, name string) (*datasourcev1.Datasource, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*datasourcev1.Datasource); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to datasourcev1.Datasource", o)
}

func (a *DatasourceLister) List(namespace string, lo metav1.ListOptions) ([]*datasourcev1.Datasource, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*datasourcev1.Datasource, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*datasourcev1.Datasource)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to datasourcev1.Datasource", o)
		}
	}
	return result, nil
}

type RestrictionLister struct {
	l *typedLister
}

func (c *Listers) RestrictionLister() *RestrictionLister {
	lister, err := c.WithResource("restrictions")
	if err != nil {
		panic(err)
	}
	return &RestrictionLister{
		lister,
	}
}

func (a *RestrictionLister) Get(namespace, name string) (*restrictionv1.Restriction, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*restrictionv1.Restriction); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to restrictionv1.Restriction", o)
}

func (a *RestrictionLister) List(namespace string, lo metav1.ListOptions) ([]*restrictionv1.Restriction, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*restrictionv1.Restriction, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*restrictionv1.Restriction)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to restrictionv1.Restriction", o)
		}
	}
	return result, nil
}

type ServiceunitLister struct {
	l *typedLister
}

func (c *Listers) ServiceunitLister() *ServiceunitLister {
	lister, err := c.WithResource("serviceunits")
	if err != nil {
		panic(err)
	}
	return &ServiceunitLister{
		lister,
	}
}

func (a *ServiceunitLister) Get(namespace, name string) (*serviceunitv1.Serviceunit, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*serviceunitv1.Serviceunit); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to serviceunitv1.Serviceunit", o)
}

func (a *ServiceunitLister) List(namespace string, lo metav1.ListOptions) ([]*serviceunitv1.Serviceunit, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*serviceunitv1.Serviceunit, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*serviceunitv1.Serviceunit)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to serviceunitv1.Serviceunit", o)
		}
	}
	return result, nil
}

type ServiceunitGroupLister struct {
	l *typedLister
}

func (c *Listers) ServiceunitGroupLister() *ServiceunitGroupLister {
	lister, err := c.WithResource("serviceunitgroups")
	if err != nil {
		panic(err)
	}
	return &ServiceunitGroupLister{
		lister,
	}
}

func (a *ServiceunitGroupLister) Get(namespace, name string) (*serviceunitgroupv1.ServiceunitGroup, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*serviceunitgroupv1.ServiceunitGroup); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to serviceunitgroupv1.ServiceunitGroup", o)
}

func (a *ServiceunitGroupLister) List(namespace string, lo metav1.ListOptions) ([]*serviceunitgroupv1.ServiceunitGroup, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*serviceunitgroupv1.ServiceunitGroup, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*serviceunitgroupv1.ServiceunitGroup)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to serviceunitgroupv1.ServiceunitGroup", o)
		}
	}
	return result, nil
}

type TopicLister struct {
	l *typedLister
}

func (c *Listers) TopicLister() *TopicLister {
	lister, err := c.WithResource("topics")
	if err != nil {
		panic(err)
	}
	return &TopicLister{
		lister,
	}
}

func (a *TopicLister) Get(namespace, name string) (*topicv1.Topic, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*topicv1.Topic); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to topicv1.Topic", o)
}

func (a *TopicLister) List(namespace string, lo metav1.ListOptions) ([]*topicv1.Topic, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*topicv1.Topic, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*topicv1.Topic)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to topicv1.Topic", o)
		}
	}
	return result, nil
}

type TopicgroupLister struct {
	l *typedLister
}

func (c *Listers) TopicgroupLister() *TopicgroupLister {
	lister, err := c.WithResource("topicgroups")
	if err != nil {
		panic(err)
	}
	return &TopicgroupLister{
		lister,
	}
}

func (a *TopicgroupLister) Get(namespace, name string) (*topicgroupv1.Topicgroup, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*topicgroupv1.Topicgroup); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to topicgroupv1.Topicgroup", o)
}

func (a *TopicgroupLister) List(namespace string, lo metav1.ListOptions) ([]*topicgroupv1.Topicgroup, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*topicgroupv1.Topicgroup, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*topicgroupv1.Topicgroup)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to topicgroupv1.Topicgroup", o)
		}
	}
	return result, nil
}

type TrafficcontrolLister struct {
	l *typedLister
}

func (c *Listers) TrafficcontrolLister() *TrafficcontrolLister {
	lister, err := c.WithResource("trafficcontrols")
	if err != nil {
		panic(err)
	}
	return &TrafficcontrolLister{
		lister,
	}
}

func (a *TrafficcontrolLister) Get(namespace, name string) (*trafficcontrolv1.Trafficcontrol, error) {
	o, err := a.l.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	if result, ok := o.(*trafficcontrolv1.Trafficcontrol); ok {
		return result, nil
	}
	return nil, fmt.Errorf("cannot cast %+v to trafficcontrolv1.Trafficcontrol", o)
}

func (a *TrafficcontrolLister) List(namespace string, lo metav1.ListOptions) ([]*trafficcontrolv1.Trafficcontrol, error) {
	list, err := a.l.List(namespace, lo)
	if err != nil {
		return nil, err
	}
	result := make([]*trafficcontrolv1.Trafficcontrol, len(list))
	for i, o := range list {
		var ok bool
		result[i], ok = o.(*trafficcontrolv1.Trafficcontrol)
		if !ok {
			return nil, fmt.Errorf("cannot cast %+v to trafficcontrolv1.Trafficcontrol", o)
		}
	}
	return result, nil
}
