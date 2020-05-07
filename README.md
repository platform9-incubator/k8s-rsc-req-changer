# k8s-rsc-req-changer
Batch modification of cpu or memory requests in a cluster

## Usage
`k8s-rsc-req-changer container-name-prefix (cpu|memory) quantity`

- If the quantity is 0, then the request is removed if present
- Otherwise, the request is modified (if already present and has diffea rent value) or added (if not present)

## Warnings
- Use with caution: all containers across all namespaces with a matching name prefix are modified
- Actually modifies deployment resources controlling the containers. Containers in pods managed by other controllers are not touched.

## Examples
```
k8s-rsc-req-changer foo- cpu 50m # changes the cpu request of all containers named foo-xxx to 50m
k8s-rsc-req-changer foo- memory 0 # removes the memory request of all containers named foo-xxx if present
```
## Building the program
Just: `go build`
