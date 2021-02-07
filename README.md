# go-parallel

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