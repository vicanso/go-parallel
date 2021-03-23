# go-parallel

[![Build Status](https://github.com/vicanso/go-parallel/workflows/Test/badge.svg)](https://github.com/vicanso/go-parallel/actions)

parallel control for go

## API

### Parallel

Runs task with limit concurrency.

```go
arr := strings.Split("0123456789", "")
fn := func(index int) error{
    fmt.Println(index)
    fmt.Println(arr[index])
    return nil
}

err := Parallel(len(arr), 3, fn)
```

### EnhancedParallel

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

### Race

Runs task parallel, when the first task is done, then return the result of it.

```go
err := Race(func() error {
    time.Sleep(time.Second)
    return errors.New("error")
}, func() error {
    // the result of this task will be used
    return nil
})
```

### Some

Runs task parallel, when the count of success task is gt count param, it will return nil. Otherwise it will return error.

```go
err := Some(5, 3, func(index int) error {
    if index%2 == 0 {
        return nil
    }
    return errors.New("error")
})
```

### Any

Runs task parallel, when one of task is success, it will return nil. Otherwise it will return error.

```go
err := Any(5, func(index int) error {
    if index == 4 {
        return nil
    }
    return errors.New("error")
})
```