module sender

go 1.12

replace (
	github.com/ymotognpoo/toolbox v0.0.0 => ../..
	github.com/ymotognpoo/toolbox/sync-tool v0.0.0 => ../
	github.com/ymotongpoo/toolbox/sync-tool/sender v0.0.0 => ./
)

require (
	github.com/rjeczalik/notify v0.9.2
	github.com/ymotongpoo/toolbox v0.0.0-20190414014614-3ae888eeeea9
	golang.org/x/oauth2 v0.0.0-20190523182746-aaccbc9213b0 // indirect
	google.golang.org/api v0.5.0 // indirect
)
