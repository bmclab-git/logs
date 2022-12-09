Set-Location ./gui
pnpm build
Set-Location ../
Write-Host "#: Updating dependencies..."
go mod tidy
$imageName = "logs"
$targetDir = "./"
$binPath = $targetDir + "/main"
Write-Host "#: building binary executable file..."
$env:GOOS="linux";$env:GOARCH="amd64";go build -ldflags="-s -w" -o $binPath ./
Write-Host "#: compressing binary executable file..."
upx $binPath
Write-Host "#: removing container..."
docker rm -f $imageName
Write-Host "#: removing docker image..."
docker rmi $imageName
Write-Host "#: building docker image"
docker build -t $imageName $targetDir
Write-Host "#: exporting docker image..."
docker save $imageName -o ./dist/$imageName.tar
Write-Host "#: copying installing script..."
Copy-Item .\$imageName.sh ./dist/$imageName.sh
Remove-Item ./main
Write-Host "#: done"