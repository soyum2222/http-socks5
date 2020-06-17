build:
	goos=windows goarch=amd64  go build -o http-socks5-win-amd64
	goos=linux goarch=amd64 go build -o http-socks5-linux-amd64
	goos=windows goarch=386 go build -o http-socks5-windows-386
	goos=linux goarch=arm go build -o http-socks5-linux-arm
