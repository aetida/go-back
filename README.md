# go-back

start

```bash
cd go-back
go mod tidy
$env:JWT_SECRET="secret_key"
go run .
```

secret_key generation

```bash
### linux
openssl rand -base64 64

### powershell
[Convert]::ToBase64String((1..64 | ForEach-Object {Get-Random -Max 256}))
```
