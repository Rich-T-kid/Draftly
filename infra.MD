# Draftly Infrastructure

## WebSocket Server
- Host IP: `3.130.118.123`  
- Docker Run Command:  
  ```bash
  docker run -d -p 7070:7070 --name draftly-ws rich329/draftly-ws:v0.0.1

## CRUD Server
- Host IP:3.13.161.172
- Docker Run Command:
    ```bash
    docker run -d -p 6060:6060 --name draftly-crud rich329/draftly-crud:v0.0.1

## Storage (S3)
- bucket: draftly-raw-documents
- region: us-east-2

## Database (Postgress)
- config: Reach out to me for config info


## API GateWay 
- TBD