### Merkle Tree

The beginnings of a merkle tree file transfer system a la bit-torrent.

The idea is to chunk a file up and store the merkle tree. Then send chunks and use the tree to verify the chunks as they arrive the other end.

```go
go run *.go -filename arrival_in_nara.txt -chunksize 8
```

The server is using (grpc)[grpc.io] as the communication layer. It is very fast and compatible with an array of different languages.

Currently there are two methods:
1. GetMetadata - Gets the metadata for the file being served.
2. GetPart     - Gets a part. A part consists of the []byte data & a merkle proof to verify the data.
