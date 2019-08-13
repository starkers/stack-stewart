# Example usage

### agent
```
RELEASE_NAME="stewart-agent"
helm \
  upgrade \
  --install \
  --namespace stewart \
  ${RELEASE_NAME} ./agent \
  --set server_address="http://fooasdasd.com/stacks
```

### server

edit the config and ensure the token from the agent is added

```
RELEASE_NAME="stewart-server"
helm \
  upgrade \
  --install \
  --namespace stewart \
  ${RELEASE_NAME} ./server/ \
  -f example-server.yaml
```
