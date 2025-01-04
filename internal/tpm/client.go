package tpm

import (
	"tpm-bunker/internal/types"
)

type Client struct {
	manager *Manager
}

func NewClient(manager *Manager) *Client {
	return &Client{
		manager: manager,
	}
}

func (c *Client) ExecuteTPMOperation(op types.UserOperation) (*types.APIResponse, error) {
	return c.manager.HandleOperation(op)
}
