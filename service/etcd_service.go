package service

import (
	"encoding/base64"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	etcd "github.com/smallnest/libkv-etcdv3-store"
)

type EtcdRegistry struct {
	kv store.Store
}

func (r *EtcdRegistry) InitRegistry() {
	etcd.Register()

	kv, err := libkv.NewStore(etcd.ETCDV3, []string{ServerConfig.RegistryURL}, nil)
	if err != nil {
		log.Printf("cannot create etcd registry: %v", err)
		return
	}
	r.kv = kv

	return
}

func (r *EtcdRegistry) FetchServices() []*Service {
	services := make(map[string]*Service)

	kvs, err := r.kv.List(ServerConfig.ServiceBaseURL)
	if err != nil {
		log.Printf("failed to list services %s: %v", ServerConfig.ServiceBaseURL, err)
		return toList(services)
	}

	for _, value := range kvs {

		nodes, err := r.kv.List(value.Key)
		if err != nil {
			log.Printf("failed to list %s: %v", value.Key, err)
			continue
		}

		for _, n := range nodes {
			key := string(n.Key[:])
			i := strings.LastIndex(key, "/")
			serviceName := strings.TrimPrefix(key[0:i], ServerConfig.ServiceBaseURL)
			var serviceAddr string
			fields := strings.Split(key, "/")
			if fields != nil && len(fields) > 1 {
				serviceAddr = fields[len(fields)-1]
			}
			v, err := url.ParseQuery(string(n.Value[:]))
			if err != nil {
				log.Println("etcd value parse failed. error: ", err.Error())
				continue
			}
			state := "n/a"
			group := ""
			if err == nil {
				state = v.Get("state")
				if state == "" {
					state = "active"
				}
				group = v.Get("group")
			}
			id := base64.StdEncoding.EncodeToString([]byte(serviceName + "@" + serviceAddr))
			service := &Service{ID: id, Name: serviceName, Address: serviceAddr, Metadata: string(n.Value[:]), State: state, Group: group}
			services[service.ID] = service
			log.Println("Service: %V", service)
		}

	}

	return toList(services)
}

func toList(m map[string]*Service) []*Service {
	var services []*Service
	for _, v := range m {
		services = append(services, v)
	}
	return services
}

func (r *EtcdRegistry) DeactivateService(name, address string) error {
	key := path.Join(ServerConfig.ServiceBaseURL, name, address)

	kv, err := r.kv.Get(key)

	if err != nil {
		return err
	}

	v, err := url.ParseQuery(string(kv.Value[:]))
	if err != nil {
		log.Println("etcd value parse failed. err ", err.Error())
		return err
	}
	v.Set("state", "inactive")
	err = r.kv.Put(kv.Key, []byte(v.Encode()), &store.WriteOptions{IsDir: false})
	if err != nil {
		log.Println("etcd set failed, err : ", err.Error())
	}

	return err
}

func (r *EtcdRegistry) ActivateService(name, address string) error {
	key := path.Join(ServerConfig.ServiceBaseURL, name, address)
	kv, err := r.kv.Get(key)

	v, err := url.ParseQuery(string(kv.Value[:]))
	if err != nil {
		log.Println("etcd value parse failed. err ", err.Error())
		return err
	}
	v.Set("state", "active")
	err = r.kv.Put(kv.Key, []byte(v.Encode()), &store.WriteOptions{IsDir: false})
	if err != nil {
		log.Println("etcdv3 put failed. err: ", err.Error())
	}

	return err
}

func (r *EtcdRegistry) UpdateMetadata(name, address string, metadata string) error {
	key := path.Join(ServerConfig.ServiceBaseURL, name, address)
	err := r.kv.Put(key, []byte(metadata), &store.WriteOptions{IsDir: false})
	return err
}
