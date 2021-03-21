# go-parallel

[![Build Status](https://github.com/vicanso/go-parallel/workflows/Test/badge.svg)](https://github.com/vicanso/go-parallel/actions)

parallel control for go

```go
arr := strings.Split("0123456789", "")
fn := func(index int) error{
    fmt.Println(index)
    fmt.Println(arr[index])
    return nil
}

err := Parallel(len(arr), 3, fn)
```

```go
arr := strings.Split("0123456789", "")
fn := func(index int) error{
    fmt.Println(index)
    fmt.Println(arr[index])
    return nil
}

err := EnhancedParallel(Option{
    Max: len(arr), 
    limit: 3,
    Task: fn,
	BreakOnError: true,
})
```