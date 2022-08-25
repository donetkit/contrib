package consul

import (
	"errors"
	servicediscovery2 "github.com/donetkit/contrib/pkg/discovery/servicediscovery"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"sync"
)

type Watcher struct {
	client   *api.Client
	option   servicediscovery2.WatchOptions
	wp       *watch.Plan
	watchers map[string]*watch.Plan
	exit     chan bool
	locker   sync.RWMutex

	next     chan *servicediscovery2.Result
	services map[string][]*servicediscovery2.Service
}

func newWatcher(client *api.Client, opts ...servicediscovery2.WatchOption) (servicediscovery2.Watcher, error) {
	var wo servicediscovery2.WatchOptions
	for _, o := range opts {
		o(&wo)
	}

	cw := &Watcher{
		option:   wo,
		client:   client,
		exit:     make(chan bool),
		next:     make(chan *servicediscovery2.Result, 10),
		watchers: make(map[string]*watch.Plan),
		services: make(map[string][]*servicediscovery2.Service),
	}

	wp, err := watch.Parse(map[string]interface{}{"type": "services"})
	if err != nil {
		return nil, err
	}

	wp.Handler = cw.handle
	go wp.RunWithClientAndHclog(client, nil)
	cw.wp = wp

	return cw, nil
}

func (cw *Watcher) Next() (*servicediscovery2.Result, error) {
	select {
	case <-cw.exit:
		return nil, errors.New("watcher stopped")
	case r, ok := <-cw.next:
		if !ok {
			return nil, errors.New("watcher stopped")
		}
		return r, nil
	}
}

func (cw *Watcher) Stop() {
	select {
	case <-cw.exit:
		return
	default:
		close(cw.exit)
		if cw.wp == nil {
			return
		}
		cw.wp.Stop()

		// drain results
		for {
			select {
			case <-cw.next:
			default:
				return
			}
		}
	}
}

func (cw *Watcher) handle(idx uint64, data interface{}) {
	services, ok := data.(map[string][]string)
	if !ok {
		return
	}
	for service, _ := range services {
		// Filter on watch options
		// wo.Service: Only watch services we care about
		if len(cw.option.Service) > 0 && service != cw.option.Service {
			continue
		}

		if _, ok := cw.watchers[service]; ok {
			continue
		}
		wp, err := watch.Parse(map[string]interface{}{
			"type":    "service",
			"service": service,
		})
		if err == nil {
			wp.Handler = cw.serviceHandler

			go wp.RunWithClientAndHclog(cw.client, nil)
			cw.watchers[service] = wp
			cw.next <- &servicediscovery2.Result{Action: "create", Service: &servicediscovery2.Service{Name: service}}
		}
	}
	cw.locker.RLock()
	// make a copy
	discoveryServices := make(map[string][]*servicediscovery2.Service)
	for k, v := range cw.services {
		discoveryServices[k] = v
	}
	cw.locker.RUnlock()

	// remove unknown services from registry
	// save the things we want to delete
	deleted := make(map[string][]*servicediscovery2.Service)

	for service, _ := range discoveryServices {
		if _, ok := services[service]; !ok {
			cw.locker.Lock()
			// save this before deleting
			deleted[service] = cw.services[service]
			delete(cw.services, service)
			cw.locker.Unlock()
		}
	}

	// remove unknown services from watchers
	for service, w := range cw.watchers {
		if _, ok := services[service]; !ok {
			w.Stop()
			delete(cw.watchers, service)
			for _, oldService := range deleted[service] {
				// send a delete for the service nodes that we're removing
				cw.next <- &servicediscovery2.Result{Action: "delete", Service: oldService}
			}
			// sent the empty list as the last resort to indicate to delete the entire service
			cw.next <- &servicediscovery2.Result{Action: "delete", Service: &servicediscovery2.Service{Name: service}}
		}
	}

}

func (cw *Watcher) serviceHandler(idx uint64, data interface{}) {
	entries, ok := data.([]*api.ServiceEntry)
	if !ok {
		return
	}
	serviceMap := map[string]*servicediscovery2.Service{}
	serviceName := ""

	for _, e := range entries {

		serviceName = e.Service.Service
		// service ID is now the node id
		id := e.Service.ID
		key := e.Service.Service

		address := e.Service.Address

		// use node address
		if len(address) == 0 {
			address = e.Node.Address
		}

		svc, ok := serviceMap[key]
		if !ok {
			svc = &servicediscovery2.Service{
				Name: e.Service.Service,
			}
			serviceMap[key] = svc
		}

		var del bool

		for _, check := range e.Checks {
			// delete the node if the status is critical
			if check.Status == "critical" {
				del = true
				break
			}
		}

		// if delete then skip the node
		if del {
			continue
		}

		svc.Nodes = append(svc.Nodes, &servicediscovery2.DefaultServiceInstance{
			Id:          id,
			ServiceName: serviceName,
			Host:        address,
			Port:        uint64(e.Service.Port),
			ClusterName: "",
			Enable:      true,
			Weight:      10,
			Healthy:     true,
			Metadata:    nil,
		})
	}

	cw.locker.RLock()
	// make a copy
	discoveryServices := make(map[string][]*servicediscovery2.Service)
	for k, v := range cw.services {
		discoveryServices[k] = v
	}
	cw.locker.RUnlock()

	var newServices []*servicediscovery2.Service

	// serviceMap is the new set of services keyed by name+version
	for _, newService := range serviceMap {
		// append to the new set of cached services
		newServices = append(newServices, newService)

		// check if the service exists in the existing cache
		oldServices, ok := discoveryServices[serviceName]
		if !ok {
			// does not exist? then we're creating brand new entries
			cw.next <- &servicediscovery2.Result{Action: "create", Service: newService}
			continue
		}

		// service exists. ok let's figure out what to update and delete version wise
		action := "create"

		for _, oldService := range oldServices {
			// does this version exist?
			// no? then default to create
			if oldService.Version != newService.Version {
				continue
			}

			// yes? then it's an update
			action = "update"

			var nodes []servicediscovery2.ServiceInstance
			// check the old nodes to see if they've been deleted
			for _, oldNode := range oldService.Nodes {
				var seen bool
				for _, newNode := range newService.Nodes {
					if newNode.GetId() == oldNode.GetId() {
						seen = true
						break
					}
				}
				// does the old node exist in the new set of nodes
				// no? then delete that shit
				if !seen {
					nodes = append(nodes, oldNode)
				}
			}

			// it's an update rather than creation
			if len(nodes) > 0 {
				delService := CopyService(oldService)
				delService.Nodes = nodes
				cw.next <- &servicediscovery2.Result{Action: "delete", Service: delService}
			}
		}

		cw.next <- &servicediscovery2.Result{Action: action, Service: newService}
	}

	// Now check old versions that may not be in new services map
	for _, old := range discoveryServices[serviceName] {
		// old version does not exist in new version map
		// kill it with fire!
		if _, ok := serviceMap[old.Version]; !ok {
			cw.next <- &servicediscovery2.Result{Action: "delete", Service: old}
		}
	}

	cw.locker.Lock()
	cw.services[serviceName] = newServices
	cw.locker.Unlock()
}

func CopyService(service *servicediscovery2.Service) *servicediscovery2.Service {
	// copy service
	s := new(servicediscovery2.Service)
	*s = *service

	// copy nodes
	nodes := make([]servicediscovery2.ServiceInstance, len(service.Nodes))
	for j, node := range service.Nodes {
		n := new(servicediscovery2.DefaultServiceInstance)
		srcNode := node.(*servicediscovery2.DefaultServiceInstance)
		*n = *srcNode
		nodes[j] = n
	}
	s.Nodes = nodes

	return s
}
