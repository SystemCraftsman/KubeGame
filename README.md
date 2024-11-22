# Developing Kubernetes Operators with Golang and Operator SDK - GoKonf 2023-2

Demo for the Developing Kubernetes Operators with Golang and Operator SDK talk at GoKonf 2023-2

## Architecture

![Architecture]()

## Examining the Operator

* Examine the ./api
* Examine the reconcilers
* Examine the custom resources and their specs
* Examine the .env

## Rough Steps

Make sure a K8S instance is running on your local.

Run the operator on your local.

Then run the following commands one by one:

```shell
k get pods -w
```

```shell
k apply -f examples/readyPlayerOne/oasis.yaml 
```

```shell
k get games
```

```shell
k get games oasis -o yaml
```

```shell
k exec -it oasis-postgres-956694c99-7gnhc -- bash
```

```shell
psql -U postgres
```

```postgresql
\c
```

```postgresql
\dt
```

```shell
k apply -f examples/readyPlayerOne/incipio.yaml
```

```postgresql
\dt
```

```postgresql
select * from world;
```

```shell
k apply -f examples/readyPlayerOne/
```

```postgresql
select * from world;
```

```shell
k delete games oasis
```