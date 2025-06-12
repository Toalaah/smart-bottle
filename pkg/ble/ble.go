package ble

import (
	"github.com/toalaah/smart-bottle/pkg/ble/client"
	"github.com/toalaah/smart-bottle/pkg/ble/service"
)

func NewService(opts ...service.ServiceOption) *service.GattService {
	return service.New(opts...)
}

func NewClient(opts ...client.ClientOption) *client.GattClient {
	return client.New(opts...)
}
