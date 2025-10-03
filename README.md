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

## Conventional Commits

Este proyecto utiliza **Conventional Commits** para versionado automático. Los commits deben seguir este formato:

```
<type>[optional scope]: <description>
```

### Tipos principales:
- `feat:` - Nueva funcionalidad (MINOR 1.0.0 → 1.1.0)
- `fix:` - Corrección de bugs (PATCH 1.0.0 → 1.0.1)
- `feat!:` o `BREAKING CHANGE:` - Cambios incompatibles (MAJOR 1.0.0 → 2.0.0)
- `docs:`, `chore:`, `refactor:`, `test:` - No incrementan versión

### Ejemplos:
```bash
git commit -m "feat: agregar endpoint /health"
git commit -m "fix: corregir validación de request"
git commit -m "docs: actualizar README"
```

## Licencia

Este proyecto está bajo la licencia MIT.