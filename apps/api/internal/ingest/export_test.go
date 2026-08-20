package ingest

// ResolveConcurrency exposes the resolver pool width to the ingest_test
// package, so the concurrency test can assert that the pool actually fills
// rather than inferring overlap from wall clock.
const ResolveConcurrency = resolveConcurrency
