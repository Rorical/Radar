package wstore

import (
	"Radar/core/utils"
	"context"
	"fmt"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dssync "github.com/ipfs/go-datastore/sync"
)

type WrappedStore struct {
	DB *dssync.MutexDatastore
}

func getInfo(key datastore.Key) {
	cid, err := utils.DecodeB32Cid(key.List()[1])
	if err != nil {
		panic(err)
	}
	fmt.Println("Detected Key:", cid)
}

func NewStore() *WrappedStore {
	db := dssync.MutexWrap(datastore.NewMapDatastore())
	return &WrappedStore{DB: db}
}

func (ws *WrappedStore) Get(ctx context.Context, key datastore.Key) (value []byte, err error) {
	getInfo(key)
	return ws.DB.Get(ctx, key)
}

func (ws *WrappedStore) Has(ctx context.Context, key datastore.Key) (exists bool, err error) {
	getInfo(key)
	return ws.DB.Has(ctx, key)
}

func (ws *WrappedStore) GetSize(ctx context.Context, key datastore.Key) (size int, err error) {
	getInfo(key)
	return ws.DB.GetSize(ctx, key)
}

func (ws *WrappedStore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	return ws.DB.Query(ctx, q)
}

func (ws *WrappedStore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	getInfo(key)
	return ws.DB.Put(ctx, key, value)
}

func (ws *WrappedStore) Delete(ctx context.Context, key datastore.Key) error {
	getInfo(key)
	return ws.DB.Delete(ctx, key)
}

func (ws *WrappedStore) Sync(ctx context.Context, prefix datastore.Key) error {
	return ws.DB.Sync(ctx, prefix)
}

func (ws *WrappedStore) Close() error {
	return ws.DB.Close()
}

func (ws *WrappedStore) Batch(ctx context.Context) (datastore.Batch, error) {
	return datastore.NewBasicBatch(ws), nil
}
