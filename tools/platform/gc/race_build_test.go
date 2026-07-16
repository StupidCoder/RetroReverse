//go:build race

package gc

// raceBuild is true when the race detector is compiled in. It is not a statement about
// threads: it is a statement about ARITHMETIC. The race build inhibits FMA contraction, and
// this machine's TEV computes a*(1-c) + b*c on every one of a million fragments a field, so
// the frame differs in its last bits from the one the pinned hashes were taken on — even
// single-threaded, where a data race cannot exist. See bench_test.go.
const raceBuild = true
