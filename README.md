# Ultra

Ultra is the backend application that I'm building to follow my performances at the gym. 

# Consul 
```bash
    docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul hashicorp/consul agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0
```

# Protoc 
```bash
    protoc -I=api --go_out=. --go-grpc_out=. health.proto
```
