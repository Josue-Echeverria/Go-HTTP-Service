# Go HTTP Service

Un servicio HTTP desarrollado en Go (Golang) que proporciona una API REST eficiente y escalable.

## Requisitos

- Docker 

### Ejecución con Docker

#### Comando para construir la imagen Docker
```bash
docker build -t go-http-service .
```

#### Comando para correr la imagen Docker
```bash
docker run --name go-http-service -p 8080:8080 go-http-service
```

## Uso

Una vez ejecutado el servicio, estará disponible en:
- URL: `http://localhost:8080`

## Licencia

Este proyecto está bajo la licencia MIT.