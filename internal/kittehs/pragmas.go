package kittehs

import "sync"

type DoNotCompare [0]func()

type DoNotCopy [0]sync.Mutex
