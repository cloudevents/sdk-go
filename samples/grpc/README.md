gRPC samples

To run the samples, you need a running gRPC server.

To run a sample gRPC server using the following command:

```bash
go run ./server/main.go
```

Then run the sender and receiver samples in separate terminals:

```bash
go run ./receiver/main.go
```

```bash
go run ./sender/main.go
```
