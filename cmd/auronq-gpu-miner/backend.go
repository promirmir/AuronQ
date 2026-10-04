package main

type gpuBackend interface {
	Name() string
	RecommendedBatch() int
	Run(initial []uint64, count int) ([]uint64, error)
	Close() error
}
