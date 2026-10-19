package handler

import "github.com/stolyarov-aa/yandex-metrics/internal/repository"

type HandlerHub struct {
	store repository.MemStorageIntf
}

func (h *HandlerHub) SetStore(store repository.MemStorageIntf) {
	h.store = store
}

func NewHandlerHub(store repository.MemStorageIntf) *HandlerHub {
	return &HandlerHub{store: store}
}
